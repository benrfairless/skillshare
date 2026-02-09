package main

import (
	"fmt"
)

func cmdAuditProject(root string, args []string) error {
	if !projectConfigExists(root) {
		return fmt.Errorf("no project config found; run 'skillshare init -p' first")
	}

	rt, err := loadProjectRuntime(root)
	if err != nil {
		return err
	}

	var specificSkill string
	for _, a := range args {
		if specificSkill == "" {
			specificSkill = a
		}
	}

	if specificSkill != "" {
		return auditSingleSkill(rt.sourcePath, specificSkill)
	}

	return auditAllSkills(rt.sourcePath)
}
