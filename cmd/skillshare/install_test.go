package main

import (
	"testing"

	"skillshare/internal/install"
)

func TestFilterSkillsByName_ExactMatch(t *testing.T) {
	skills := []install.SkillInfo{
		{Name: "figma", Path: "figma"},
		{Name: "pdf", Path: "pdf"},
		{Name: "github-actions", Path: "github-actions"},
	}

	matched, notFound := filterSkillsByName(skills, []string{"pdf"})
	if len(notFound) > 0 {
		t.Errorf("unexpected notFound: %v", notFound)
	}
	if len(matched) != 1 || matched[0].Name != "pdf" {
		t.Errorf("expected [pdf], got %v", matched)
	}
}

func TestFilterSkillsByName_FuzzyMatchSingle(t *testing.T) {
	skills := []install.SkillInfo{
		{Name: "figma", Path: "figma"},
		{Name: "pdf", Path: "pdf"},
		{Name: "github-actions", Path: "github-actions"},
	}

	matched, notFound := filterSkillsByName(skills, []string{"fig"})
	if len(notFound) > 0 {
		t.Errorf("unexpected notFound: %v", notFound)
	}
	if len(matched) != 1 || matched[0].Name != "figma" {
		t.Errorf("expected fuzzy match [figma], got %v", matched)
	}
}

func TestFilterSkillsByName_FuzzyMatchAmbiguous(t *testing.T) {
	skills := []install.SkillInfo{
		{Name: "github-actions", Path: "github-actions"},
		{Name: "github-pr", Path: "github-pr"},
		{Name: "pdf", Path: "pdf"},
	}

	matched, notFound := filterSkillsByName(skills, []string{"github"})
	if len(matched) != 0 {
		t.Errorf("expected no match for ambiguous query, got %v", matched)
	}
	if len(notFound) != 1 {
		t.Fatalf("expected 1 notFound, got %d", len(notFound))
	}
	if notFound[0] == "github" {
		t.Error("notFound should contain 'did you mean' suggestions, got plain name")
	}
}

func TestFilterSkillsByName_NotFound(t *testing.T) {
	skills := []install.SkillInfo{
		{Name: "figma", Path: "figma"},
		{Name: "pdf", Path: "pdf"},
	}

	matched, notFound := filterSkillsByName(skills, []string{"zzzzz"})
	if len(matched) != 0 {
		t.Errorf("expected no matches, got %v", matched)
	}
	if len(notFound) != 1 || notFound[0] != "zzzzz" {
		t.Errorf("expected notFound=[zzzzz], got %v", notFound)
	}
}

func TestFilterSkillsByName_ExactTakesPriority(t *testing.T) {
	skills := []install.SkillInfo{
		{Name: "fig", Path: "fig"},
		{Name: "figma", Path: "figma"},
	}

	matched, notFound := filterSkillsByName(skills, []string{"fig"})
	if len(notFound) > 0 {
		t.Errorf("unexpected notFound: %v", notFound)
	}
	if len(matched) != 1 || matched[0].Name != "fig" {
		t.Errorf("exact match should take priority, got %v", matched)
	}
}
