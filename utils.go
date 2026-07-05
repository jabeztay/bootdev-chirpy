package main

import "regexp"

func removeProfane(text string) string {
	re := regexp.MustCompile(`(?i)kerfuffle|sharbert|fornax`)
	result := re.ReplaceAllString(text, "****")
	return result
}
