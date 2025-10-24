package models

import (
	"encoding/json"
)

type WgerExercise struct {
	ID               int             `json:"id"`
	UUID             string          `json:"uuid"`
	Name             string          `json:"name"`
	Description      string          `json:"description"`
	Category         WgerCategory    `json:"category"`
	Muscles          []WgerMuscle    `json:"muscles"`
	MusclesSecondary []WgerMuscle    `json:"muscles_secondary"`
	Equipment        []WgerEquipment `json:"equipment"`
	Language         int             `json:"language"`
	License          WgerLicense     `json:"license"`
	LicenseAuthor    string          `json:"license_author"`
	Variations       json.RawMessage `json:"variations"`
	CreationDate     string          `json:"creation_date"`
}

type WgerCategory struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type WgerLicense struct {
	ID        int    `json:"id"`
	FullName  string `json:"full_name"`
	ShortName string `json:"short_name"`
	URL       string `json:"url"`
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

type WgerCategoryListResponse struct {
	Count    int            `json:"count"`
	Next     *string        `json:"next"`
	Previous *string        `json:"previous"`
	Results  []WgerCategory `json:"results"`
}

func (w *WgerExercise) GetVariations() []int {
	if w.Variations == nil {
		return nil
	}

	var variations []int
	if err := json.Unmarshal(w.Variations, &variations); err == nil {
		return variations
	}

	var singleVariation int
	if err := json.Unmarshal(w.Variations, &singleVariation); err == nil {
		return []int{singleVariation}
	}

	return nil
}
