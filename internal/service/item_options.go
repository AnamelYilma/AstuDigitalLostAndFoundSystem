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

var itemTypes = []string{
	"lost",
	"found",
}

var itemCategories = []string{
	
	"electronics",
	"book & document",
	"study tool",
	"atm card",
	"jewelry",
	"sport equipment",
	"bag & backpack",
	"key",
	"id card",
	"clothing & accessories",
	"other",
}

var approvalStatuses = []string{
	"pending",
	"approved",
	"rejected",
}

func ASTULocations() []string {
	return astuLocations
}

func ColorOptions() []string {
	return colorOptions
}

func ItemCategories() []string {
	return itemCategories
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

func IsValidItemType(itemType string) bool {
	for _, t := range itemTypes {
		if strings.EqualFold(strings.TrimSpace(itemType), t) {
			return true
		}
	}
	return false
}

func IsValidCategory(category string) bool {
	for _, c := range itemCategories {
		if strings.EqualFold(strings.TrimSpace(category), c) {
			return true
		}
	}
	return false
}

func IsValidApprovalStatus(status string) bool {
	for _, s := range approvalStatuses {
		if strings.EqualFold(strings.TrimSpace(status), s) {
			return true
		}
	}
	return false
}

func IsValidASTULocation(location string) bool {
	for _, l := range astuLocations {
		if strings.EqualFold(strings.TrimSpace(location), l) {
			return true
		}
	}
	return false
}
