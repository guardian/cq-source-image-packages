package client

type Spec struct {
	AmigoBucketName          string `json:"bucket"`
	AmigoBakesTableName      string `json:"bakes_table"`
	AmigoRecipesTableName    string `json:"recipes_table"`
	AmigoBaseImagesTableName string `json:"base_images_table"`
}
