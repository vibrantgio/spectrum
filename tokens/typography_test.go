package tokens_test

import (
	"testing"

	"github.com/vibrantgio/spectrum/tokens"
)

func TestDefaultTypographyRolesComplete(t *testing.T) {
	roles := []struct {
		name  string
		style tokens.TextStyle
	}{
		{"DisplayLarge", tokens.DefaultTypography.DisplayLarge},
		{"DisplayMedium", tokens.DefaultTypography.DisplayMedium},
		{"DisplaySmall", tokens.DefaultTypography.DisplaySmall},
		{"HeadlineLarge", tokens.DefaultTypography.HeadlineLarge},
		{"HeadlineMedium", tokens.DefaultTypography.HeadlineMedium},
		{"HeadlineSmall", tokens.DefaultTypography.HeadlineSmall},
		{"TitleLarge", tokens.DefaultTypography.TitleLarge},
		{"TitleMedium", tokens.DefaultTypography.TitleMedium},
		{"TitleSmall", tokens.DefaultTypography.TitleSmall},
		{"LabelLarge", tokens.DefaultTypography.LabelLarge},
		{"LabelMedium", tokens.DefaultTypography.LabelMedium},
		{"LabelSmall", tokens.DefaultTypography.LabelSmall},
		{"BodyLarge", tokens.DefaultTypography.BodyLarge},
		{"BodyMedium", tokens.DefaultTypography.BodyMedium},
		{"BodySmall", tokens.DefaultTypography.BodySmall},
	}
	for _, role := range roles {
		if role.style.Size <= 0 {
			t.Errorf("%s: zero size", role.name)
		}
		if role.style.Weight <= 0 {
			t.Errorf("%s: zero weight", role.name)
		}
		if role.style.LineHeight <= 0 {
			t.Errorf("%s: zero line height", role.name)
		}
	}
}
