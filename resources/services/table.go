package services

import (
	"context"
	"fmt"
	"github.com/cloudquery/plugin-sdk/v4/schema"
	"github.com/cloudquery/plugin-sdk/v4/transformers"
	"github.com/guardian/cq-source-images-instances/client"
	"regexp"
	"strings"
)

type AmigoBakePackage struct {
	ImageId        string `json:"image_id"`
	BakeId         string `json:"bake_id"`
	PackageName    string `json:"package_name"`
	PackageVersion string `json:"package_version"`
}

func AmigoBakePackages() *schema.Table {
	return &schema.Table{
		Name:      "amigo_bake_packages",
		Resolver:  fetchAmigoBakePackages,
		Transform: transformers.TransformWithStruct(AmigoBakePackage{}),
	}
}

func fetchAmigoBakePackages(ctx context.Context, meta schema.ClientMeta, parent *schema.Resource, res chan<- any) error {
	cl := meta.(*client.Client)
	store := cl.Store

	keys, err := store.ListKeys()
	if err != nil {
		return err
	}

	records := map[string]AmigoBakePackage{}
	for _, key := range keys {
		imageId, bakeId, err := extractImageAndBakeIds(key)
		if err == nil {
			data, err := store.Get(key)
			if err != nil {
				return err
			}
			bakePackages, err := unmarshalSpaceSeparated(imageId, bakeId, data)
			if err != nil {
				return err
			}
			records = joinMaps(records, bakePackages)
		}
	}

	for _, record := range records {
		res <- record
	}

	return nil
}

func extractImageAndBakeIds(key string) (string, string, error) {
	parts := strings.SplitN(strings.TrimSuffix(key, ".txt"), "--", 2)
	if len(parts) == 2 {
		return parts[0], parts[1], nil
	} else {
		return "", "", fmt.Errorf("invalid key format: %s", key)
	}
}

func unmarshalSpaceSeparated(imageId string, bakeId string, data []byte) (map[string]AmigoBakePackage, error) {
	lines := strings.Split(string(data), "\n")

	fieldSplitter := regexp.MustCompile(`\s+`)

	records := make(map[string]AmigoBakePackage)

	for _, line := range lines {
		fields := fieldSplitter.Split(line, 2)
		if len(fields) == 2 {
			record := AmigoBakePackage{
				ImageId:        imageId,
				BakeId:         bakeId,
				PackageName:    fields[0],
				PackageVersion: fields[1],
			}
			records[fmt.Sprintf("%s_%s_%s", record.ImageId, record.BakeId, record.PackageName)] = record
		}
	}

	fmt.Println("records:", records)

	return records, nil
}

func joinMaps(map1, map2 map[string]AmigoBakePackage) map[string]AmigoBakePackage {
	for key, value := range map2 {
		map1[key] = value
	}
	return map1
}
