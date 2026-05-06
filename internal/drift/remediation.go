package drift

import (
	"fmt"
	"io"
	"strings"
	"time"
)

// RemediationAction describes a suggested fix for a drift entry.
type RemediationAction struct {
	Kind        DriftKind `yaml:"kind"`
	Service     string    `yaml:"service"`
	Description string    `yaml:"description"`
	Command     string    `yaml:"command"`
	CreatedAt   time.Time `yaml:"created_at"`
}

// RemediationPlan holds a set of actions for a drifted service.
type RemediationPlan struct {
	Service string
	Actions []RemediationAction
}

// HasActions returns true when the plan contains at least one action.
func (p *RemediationPlan) HasActions() bool {
	return len(p.Actions) > 0
}

// WriteTo writes the plan in human-readable form to w.
func (p *RemediationPlan) WriteTo(w io.Writer) (int64, error) {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Remediation plan for service: %s\n", p.Service))
	if !p.HasActions() {
		sb.WriteString("  No actions required.\n")
	} else {
		for i, a := range p.Actions {
			sb.WriteString(fmt.Sprintf("  [%d] (%s) %s\n", i+1, a.Kind, a.Description))
			if a.Command != "" {
				sb.WriteString(fmt.Sprintf("      $ %s\n", a.Command))
			}
		}
	}
	n, err := fmt.Fprint(w, sb.String())
	return int64(n), err
}

// BuildRemediationPlan produces a RemediationPlan from a DriftResult.
func BuildRemediationPlan(result DriftResult) RemediationPlan {
	plan := RemediationPlan{Service: result.Service}
	for _, e := range result.Entries {
		action := remediationFor(result.Service, e)
		plan.Actions = append(plan.Actions, action)
	}
	return plan
}

func remediationFor(service string, e DriftEntry) RemediationAction {
	a := RemediationAction{
		Kind:      e.Kind,
		Service:   service,
		CreatedAt: time.Now().UTC(),
	}
	switch e.Kind {
	case DriftKindImage:
		a.Description = fmt.Sprintf("Update image from %q to %q", e.Got, e.Want)
		a.Command = fmt.Sprintf("kubectl set image deployment/%s app=%s", service, e.Want)
	case DriftKindReplicas:
		a.Description = fmt.Sprintf("Scale replicas from %v to %v", e.Got, e.Want)
		a.Command = fmt.Sprintf("kubectl scale deployment/%s --replicas=%v", service, e.Want)
	case DriftKindEnv:
		a.Description = fmt.Sprintf("Reconcile env var %q (got %v, want %v)", e.Field, e.Got, e.Want)
		a.Command = ""
	default:
		a.Description = fmt.Sprintf("Reconcile field %q (got %v, want %v)", e.Field, e.Got, e.Want)
	}
	return a
}
