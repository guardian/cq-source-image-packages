package services

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/cloudquery/plugin-sdk/v4/schema"
	"github.com/cloudquery/plugin-sdk/v4/transformers"
	"github.com/guardian/cq-source-image-packages/client"
	"github.com/guardian/cq-source-image-packages/store"
)

type AmigoBakePackage struct {
	BaseName    string    `json:"base_name"`
	BaseAmiId   string    `json:"base_ami_id"` // The 'SourceAMI' field in Amiable
	BaseEolDate time.Time `json:"base_eol_date"`
	RecipeId    string    `json:"recipe_id"`
	BakeId      string    `json:"bake_id"`
	SourceAmiId string    `json:"ami_id"` // The 'CopiedFromAMI' field in Amiable
	//TODO AwsAccountId   string    `json:"aws_account_id"`
	AwsAccountIds  []string  `json:"aws_account_ids"`
	StartedAt      time.Time `json:"started_at"`
	StartedBy      string    `json:"started_by"`
	PackageName    string    `json:"package_name"`
	PackageVersion string    `json:"package_version"`
}

type AmigoRecipe struct {
	RecipeId      string
	BaseName      string
	AwsAccountIds []string
}

type AmigoBaseImage struct {
	BaseName    string
	BaseAmiId   string
	BaseEolDate time.Time
}

type OsPackage struct {
	PackageName    string
	PackageVersion string
}

func AmigoBakePackages() *schema.Table {
	return &schema.Table{
		Name:      "amigo_bake_packages",
		Resolver:  fetchAmigoBakePackages,
		Transform: transformers.TransformWithStruct(AmigoBakePackage{}),
	}
}

func fetchAmigoBakePackages(_ context.Context, meta schema.ClientMeta, _ *schema.Resource, res chan<- any) error {
	cl := meta.(*client.Client)
	s3Store := cl.S3Store
	bakesTable := cl.BakesTable
	recipesTable := cl.RecipesTable
	baseImagesTable := cl.BaseImagesTable

	allBakes := bakesTable.ListAll()

	allRecipes := map[string]AmigoRecipe{}
	for _, recipe := range recipesTable.ListAll() {
		id := recipe["id"].(*types.AttributeValueMemberS).Value
		baseName := recipe["baseImageId"].(*types.AttributeValueMemberS).Value

		var accountIds []string
		if accountIdsAttrib, ok := recipe["encryptFor"]; ok {
			accountIds = attributeValuesToStrings(accountIdsAttrib.(*types.AttributeValueMemberL).Value)
		} else {
			log.Printf("Recipe '%v' has no 'encryptFor' attribute\n", id)
			accountIds = []string{}
		}

		allRecipes[id] = AmigoRecipe{
			RecipeId:      id,
			BaseName:      baseName,
			AwsAccountIds: accountIds,
		}
	}

	allBaseImages := map[string]AmigoBaseImage{}
	for _, baseImg := range baseImagesTable.ListAll() {
		id := baseImg["id"].(*types.AttributeValueMemberS).Value
		baseEolDate := baseImg["eolDate"].(*types.AttributeValueMemberS).Value
		img := AmigoBaseImage{
			BaseName:  id,
			BaseAmiId: baseImg["amiId"].(*types.AttributeValueMemberS).Value,
		}
		img.SetBaseEolDate(baseEolDate)
		allBaseImages[id] = img
	}

	records := map[string]AmigoBakePackage{}
	for _, bake := range allBakes {
		recipeId := bake["recipeId"].(*types.AttributeValueMemberS).Value
		bakeId := bake["buildNumber"].(*types.AttributeValueMemberN).Value
		status := bake["status"].(*types.AttributeValueMemberS).Value
		startedAt := bake["startedAt"].(*types.AttributeValueMemberS).Value
		startedBy := bake["startedBy"].(*types.AttributeValueMemberS).Value

		// Skip if bake doesn't have an AMI ID
		var amiId string
		if amiIdAttrib, ok := bake["amiId"]; ok {
			amiId = amiIdAttrib.(*types.AttributeValueMemberS).Value
		} else {
			log.Printf("Skipping bake without AMI ID: recipeId: '%v', bakeId: '%v', status: '%v'\n", recipeId, bakeId, status)
			continue
		}

		recipe := allRecipes[recipeId]
		baseImage := allBaseImages[recipe.BaseName]

		// Fetch corresponding package file from S3
		packages, err := fetchBakePackages(recipeId, bakeId, s3Store)
		if err != nil {
			log.Fatalf("Error fetching bake packages for recipe '%s', bake '%s': %v", recipeId, bakeId, err)
		}

		// Create a record for each package
		for _, pkg := range packages {
			key := fmt.Sprintf("%s--%s--%s--%s", recipeId, bakeId, amiId, pkg.PackageName)
			bakePackage := AmigoBakePackage{
				BaseName:       recipe.BaseName,
				BaseAmiId:      baseImage.BaseAmiId,
				BaseEolDate:    baseImage.BaseEolDate,
				RecipeId:       recipeId,
				BakeId:         bakeId,
				SourceAmiId:    amiId,
				AwsAccountIds:  recipe.AwsAccountIds,
				StartedBy:      startedBy,
				PackageName:    pkg.PackageName,
				PackageVersion: pkg.PackageVersion,
			}
			bakePackage.SetStartedAt(startedAt)
			records[key] = bakePackage
		}
	}

	for _, record := range records {
		res <- record
	}

	return nil
}

func toTime(s string) (time.Time, error) {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		log.Printf("Failed to parse time: %v", err)
	}
	return t, err
}

func (a *AmigoBaseImage) SetBaseEolDate(s string) {
	t, err := toTime(s)
	if err == nil {
		a.BaseEolDate = t
	}
}

func (a *AmigoBakePackage) SetStartedAt(s string) {
	t, err := toTime(s)
	if err == nil {
		a.StartedAt = t
	}
}

func attributeValuesToStrings(avList []types.AttributeValue) []string {
	var strList []string
	for _, av := range avList {
		if avMemberS, ok := av.(*types.AttributeValueMemberS); ok {
			strList = append(strList, avMemberS.Value)
		}
	}
	return strList
}

// given a recipe ID and bake number, fetch corresponding packages file from S3
// and extract its contents into a list of OsPackages.
func fetchBakePackages(recipeId string, bakeId string, s3Store store.S3) ([]OsPackage, error) {
	key := fmt.Sprintf("%s--%s.txt", recipeId, bakeId)
	data, err := s3Store.Get(key)

	if err != nil {
		// If file doesn't exist, return an empty slice
		if strings.Contains(err.Error(), "NoSuchKey") {
			return []OsPackage{}, nil
		}
		return nil, err
	}

	return parseSpaceSeparated(key, data), nil
}

// parseSpaceSeparated parses a space-separated list of package names and versions
func parseSpaceSeparated(listName string, packageData []byte) []OsPackage {
	fieldSplitter := regexp.MustCompile(`\s+`)
	lines := strings.Split(string(packageData), "\n")

	var packages []OsPackage
	for i, line := range lines {
		pkg, err := parseLine(line, fieldSplitter)
		if err != nil {
			log.Printf("List %v line %d skipped: %v", listName, i+1, err)
			continue
		}
		packages = append(packages, pkg)
	}

	return packages
}

func parseLine(line string, fieldSplitter *regexp.Regexp) (OsPackage, error) {
	fields := fieldSplitter.Split(line, 2)
	if len(fields) != 2 {
		return OsPackage{}, fmt.Errorf("invalid package line: '%s'", line)
	}
	return OsPackage{
		PackageName:    fields[0],
		PackageVersion: fields[1],
	}, nil
}
