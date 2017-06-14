package uses

import (
	"fmt"
	"path/filepath"
	"strings"
)

func Why(index *Index, mainTarget string) (reasons []string) {
	mainTarget = trimExt(filepath.Base(mainTarget))
	fmt.Println(mainTarget)

	usedBy := make(map[string][]string, len(index.Uses))
	for path, uses := range index.Uses {
		path = strings.ToLower(path)
		for _, target := range uses.Interface {
			target = strings.ToLower(target)
			usedBy[target] = includeString(usedBy[target], path)
		}
		for _, target := range uses.Implementation {
			target = strings.ToLower(target)
			usedBy[target] = includeString(usedBy[target], path)
		}
	}

	roots := make(map[string]bool)
	for _, root := range index.RootFiles {
		roots[strings.ToLower(root)] = true
	}

	var reverse func(target string, chain []string)
	reverse = func(target string, chain []string) {
		if roots[target] {
			reason := ""
			for i := len(chain) - 1; i >= 0; i-- {
				reason += ">" + chain[i]
			}
			fmt.Println(reason)
			reasons = append(reasons, reason)
			return
		}
		if contains(target, chain) {
			return
		}

		chain = append(chain, target)
		for _, parent := range usedBy[target] {
			reverse(parent, chain)
		}
	}
	reverse(strings.ToLower(mainTarget), nil)

	return reasons
}
