package models

type WgerExercise struct {
	ID               int    `json:"id"`
	UUID             string `json:"uuid"`
	Name             string `json:"name"`
	Description      string `json:"description"`
	Category         int    `json:"category"`
	Muscles          []int  `json:"muscles"`
	MusclesSecondary []int  `json:"muscles_secondary"`
	Equipment        []int  `json:"equipment"`
	Language         int    `json:"language"`
	License          int    `json:"license"`
	LicenseAuthor    string `json:"license_author"`
	Variations       []int  `json:"variations"`
	CreationDate     string `json:"creation_date"`
}

type WgerExerciseListResponse struct {
	Count    int            `json:"count"`
	Next     *string        `json:"next"`
	Previous *string        `json:"previous"`
	Results  []WgerExercise `json:"results"`
}

type WgerMuscle struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	NameEn  string `json:"name_en"`
	IsFront bool   `json:"is_front"`
}

type WgerMuscleListResponse struct {
	Count    int          `json:"count"`
	Next     *string      `json:"next"`
	Previous *string      `json:"previous"`
	Results  []WgerMuscle `json:"results"`
}

type WgerEquipment struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type WgerEquipmentListResponse struct {
	Count    int             `json:"count"`
	Next     *string         `json:"next"`
	Previous *string         `json:"previous"`
	Results  []WgerEquipment `json:"results"`
}

type WgerCategory struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type WgerCategoryListResponse struct {
	Count    int            `json:"count"`
	Next     *string        `json:"next"`
	Previous *string        `json:"previous"`
	Results  []WgerCategory `json:"results"`
}
