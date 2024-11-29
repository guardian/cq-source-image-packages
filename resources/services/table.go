package services

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/cloudquery/plugin-sdk/v4/schema"
	"github.com/cloudquery/plugin-sdk/v4/transformers"
	"github.com/guardian/cq-source-image-packages/client"
	"github.com/guardian/cq-source-image-packages/store"
)

type AmigoBakePackage struct {
	BaseName       string    `json:"base_name"`
	BaseAmiId      string    `json:"base_ami_id"` // The 'SourceAMI' field in Amiable
	BaseEolDate    time.Time `json:"base_eol_date"`
	RecipeId       string    `json:"recipe_id"`
	BakeNumber     int       `json:"bake_number"`
	SourceAmiId    string    `json:"source_ami_id"` // The 'CopiedFromAMI' field in Amiable
	StartedAt      time.Time `json:"started_at"`
	StartedBy      string    `json:"started_by"`
	PackageName    string    `json:"package_name"`
	PackageVersion string    `json:"package_version"`
}

type AmigoRecipe struct {
	RecipeId string
	BaseName string
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
		Resolver:  FetchAmigoBakePackages,
		Transform: transformers.TransformWithStruct(AmigoBakePackage{}),
	}
}

func FetchAmigoBakePackages(_ context.Context, meta schema.ClientMeta, _ *schema.Resource, res chan<- any) error {
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
		allRecipes[id] = AmigoRecipe{
			RecipeId: id,
			BaseName: baseName,
		}
	}

	allBaseImages := map[string]AmigoBaseImage{}
	for _, baseImgAttribs := range baseImagesTable.ListAll() {
		baseImage := toAmigoBaseImage(baseImgAttribs)
		allBaseImages[baseImage.BaseName] = baseImage
	}

	records := map[string]AmigoBakePackage{}
	for _, bake := range allBakes {
		recipeId := bake["recipeId"].(*types.AttributeValueMemberS).Value
		bakeNumber, err := strconv.Atoi(bake["buildNumber"].(*types.AttributeValueMemberN).Value)
		if err != nil {
			log.Printf("Failed to convert buildNumber to int: %v", err)
			continue
		}
		status := bake["status"].(*types.AttributeValueMemberS).Value
		startedAt := bake["startedAt"].(*types.AttributeValueMemberS).Value
		startedBy := bake["startedBy"].(*types.AttributeValueMemberS).Value

		// Skip if bake doesn't have an AMI ID
		var amiId string
		if amiIdAttrib, ok := bake["amiId"]; ok {
			amiId = amiIdAttrib.(*types.AttributeValueMemberS).Value
		} else {
			log.Printf("Skipping bake without AMI ID: recipeId: '%v', bakeId: '%v', status: '%v'\n", recipeId, bakeNumber, status)
			continue
		}

		recipe := allRecipes[recipeId]
		baseImage := allBaseImages[recipe.BaseName]

		// Fetch corresponding package file from S3
		packages, err := fetchBakePackages(recipeId, bakeNumber, s3Store)
		if err != nil {
			log.Fatalf("Error fetching bake packages for recipe '%s', bake '%d': %v", recipeId, bakeNumber, err)
		}

		// Create a record for each package
		for _, pkg := range packages {
			key := fmt.Sprintf("%s--%d--%s--%s", recipeId, bakeNumber, amiId, pkg.PackageName)
			bakePackage := AmigoBakePackage{
				BaseName:       recipe.BaseName,
				BaseAmiId:      baseImage.BaseAmiId,
				BaseEolDate:    baseImage.BaseEolDate,
				RecipeId:       recipeId,
				BakeNumber:     bakeNumber,
				SourceAmiId:    amiId,
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

func toAmigoBaseImage(baseImgAttribs map[string]types.AttributeValue) AmigoBaseImage {
	baseImage := AmigoBaseImage{
		BaseName:  baseImgAttribs["id"].(*types.AttributeValueMemberS).Value,
		BaseAmiId: baseImgAttribs["amiId"].(*types.AttributeValueMemberS).Value,
	}
	if eolDateAttrib, ok := baseImgAttribs["eolDate"]; ok && eolDateAttrib != nil {
		baseEolDate := eolDateAttrib.(*types.AttributeValueMemberS).Value
		baseImage.SetBaseEolDate(baseEolDate)
	} else {
		baseImage.BaseEolDate = time.Time{}
	}
	return baseImage
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

// given a recipe ID and bake number, fetch corresponding packages file from S3
// and extract its contents into a list of OsPackages.
func fetchBakePackages(recipeId string, bakeNumber int, s3Store store.S3) ([]OsPackage, error) {
	key := fmt.Sprintf("%s--%d.txt", recipeId, bakeNumber)
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
