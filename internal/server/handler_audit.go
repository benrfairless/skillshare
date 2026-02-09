package server

import (
	"net/http"
	"os"
	"path/filepath"

	"skillshare/internal/audit"
	"skillshare/internal/sync"
	"skillshare/internal/utils"
)

type auditFindingResponse struct {
	Severity string `json:"severity"`
	Pattern  string `json:"pattern"`
	Message  string `json:"message"`
	File     string `json:"file"`
	Line     int    `json:"line"`
	Snippet  string `json:"snippet"`
}

type auditResultResponse struct {
	SkillName string                 `json:"skillName"`
	Findings  []auditFindingResponse `json:"findings"`
}

type auditSummary struct {
	Total   int `json:"total"`
	Passed  int `json:"passed"`
	Warning int `json:"warning"`
	Failed  int `json:"failed"`
}

func (s *Server) handleAuditAll(w http.ResponseWriter, r *http.Request) {
	source := s.cfg.Source

	// Discover all skills
	discovered, err := sync.DiscoverSourceSkills(source)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Deduplicate and also pick up top-level dirs without SKILL.md
	seen := make(map[string]bool)
	type skillEntry struct {
		name string
		path string
	}
	var skills []skillEntry

	for _, d := range discovered {
		if seen[d.SourcePath] {
			continue
		}
		seen[d.SourcePath] = true
		skills = append(skills, skillEntry{d.FlatName, d.SourcePath})
	}

	entries, _ := os.ReadDir(source)
	for _, e := range entries {
		if !e.IsDir() || utils.IsHidden(e.Name()) {
			continue
		}
		p := filepath.Join(source, e.Name())
		if !seen[p] {
			seen[p] = true
			skills = append(skills, skillEntry{e.Name(), p})
		}
	}

	var results []auditResultResponse
	summary := auditSummary{Total: len(skills)}

	for _, sk := range skills {
		result, err := audit.ScanSkill(sk.path)
		if err != nil {
			continue
		}

		resp := toAuditResponse(result)
		results = append(results, resp)

		switch result.MaxSeverity() {
		case audit.SeverityCritical, audit.SeverityHigh:
			summary.Failed++
		case audit.SeverityMedium:
			summary.Warning++
		default:
			summary.Passed++
		}
	}

	writeJSON(w, map[string]any{
		"results": results,
		"summary": summary,
	})
}

func (s *Server) handleAuditSkill(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	skillPath := filepath.Join(s.cfg.Source, name)

	if _, err := os.Stat(skillPath); os.IsNotExist(err) {
		writeError(w, http.StatusNotFound, "skill not found: "+name)
		return
	}

	result, err := audit.ScanSkill(skillPath)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, toAuditResponse(result))
}

func toAuditResponse(result *audit.Result) auditResultResponse {
	findings := make([]auditFindingResponse, 0, len(result.Findings))
	for _, f := range result.Findings {
		findings = append(findings, auditFindingResponse{
			Severity: f.Severity,
			Pattern:  f.Pattern,
			Message:  f.Message,
			File:     f.File,
			Line:     f.Line,
			Snippet:  f.Snippet,
		})
	}
	return auditResultResponse{
		SkillName: result.SkillName,
		Findings:  findings,
	}
}
