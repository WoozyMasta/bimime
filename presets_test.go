package bimime

import "testing"

func TestBIAmbiguousRAPExtensionsIsImmutable(t *testing.T) {
	t.Parallel()

	extensions := BIAmbiguousRAPExtensions()
	if len(extensions) != 4 {
		t.Fatalf("len(extensions)=%d want 4", len(extensions))
	}

	extensions[0] = "changed"
	again := BIAmbiguousRAPExtensions()
	if again[0] == "changed" {
		t.Fatal("BIAmbiguousRAPExtensions must return independent copy")
	}
}

func TestBIAmbiguousRAPOverrides(t *testing.T) {
	t.Parallel()

	plans := BIAmbiguousRAPOverrides()
	if len(plans) != 4 {
		t.Fatalf("len(plans)=%d want 4", len(plans))
	}

	for _, ext := range []string{"p3d", "wrp", "rvmat", "bisurf"} {
		plan, ok := plans[ext]
		if !ok {
			t.Fatalf("missing plan for extension %q", ext)
		}
		if plan.Match != AnalyzeMatchExtensionMagic {
			t.Fatalf("plan.Match(%s)=%v want %v", ext, plan.Match, AnalyzeMatchExtensionMagic)
		}
		if plan.Validate != AnalyzeValidateNone {
			t.Fatalf("plan.Validate(%s)=%v want %v", ext, plan.Validate, AnalyzeValidateNone)
		}
	}
}

func TestBIAmbiguousRAPOptions(t *testing.T) {
	t.Parallel()

	options := BIAmbiguousRAPOptions("terrain.wrp", []byte("9VBW"))
	if options.Path != "terrain.wrp" {
		t.Fatalf("Path=%q want %q", options.Path, "terrain.wrp")
	}
	if string(options.Prefix) != "9VBW" {
		t.Fatalf("Prefix=%q want %q", string(options.Prefix), "9VBW")
	}
	if options.DefaultPlan != PlanFast() {
		t.Fatalf("DefaultPlan=%v want %v", options.DefaultPlan, PlanFast())
	}
	if len(options.PlansByExtension) != 4 {
		t.Fatalf("len(PlansByExtension)=%d want 4", len(options.PlansByExtension))
	}
}
