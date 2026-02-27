package service

import "strings"

var astuLocations = []string{
	"Library",
	"Cafe",
	"Class",
	"Lap",
	"Dorm",
	"On Road",
	"Tolest",
	"Shower",
	"Anphe",
	"Launch",
	"Park",
	"Hale.Birroe",
	"Other",
}

var colorOptions = []string{
	"red",
	"green",
	"blue",
	"yellow",
	"black",
	"white",
	"gray",
	"brown",
	"orange",
	"purple",
	"pink",
	"gold",
	"silver",
	"other",
}

func ASTULocations() []string {
	return astuLocations
}

func ColorOptions() []string {
	return colorOptions
}

func IsStandardColor(color string) bool {
	for _, c := range colorOptions {
		if strings.EqualFold(strings.TrimSpace(color), c) {
			return true
		}
	}
	return false
}

func KnownNonOtherColors() []string {
	return []string{"red", "green", "blue", "yellow", "black", "white", "gray", "brown", "orange", "purple", "pink", "gold", "silver"}
}

func IsValidASTULocation(location string) bool {
	for _, l := range astuLocations {
		if strings.EqualFold(strings.TrimSpace(location), l) {
			return true
		}
	}
	return false
}
