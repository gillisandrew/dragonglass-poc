package main

import (
	"context"
	"fmt"
	"os"

	"dagger.io/dagger/dag"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Usage: %s <repository> <ref>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Example: %s https://github.com/gillisandrew/dragonglass-poc.git main\n", os.Args[0])
		os.Exit(1)
	}

	repository := os.Args[1]
	ref := os.Args[2]

	if err := build(context.Background(), repository, ref); err != nil {
		fmt.Println(err)
	}
}

func build(ctx context.Context, repository, ref string) error {
	fmt.Println("Building with Dagger")
	defer dag.Close()

	// create empty directory to put build outputs
	outputs := dag.Directory()
	repo := dag.Git(repository).Ref(ref).Tree()
	installer := dag.Container().
		From("node:22").
		WithDirectory("/usr/src/plugin", repo.Directory("example-plugin")).
		WithWorkdir("/usr/src/plugin").
		WithExec([]string{"npm", "ci"}).
		WithExec([]string{"bash", "-c", "npm sbom --sbom-type application --sbom-format spdx > sbom.spdx.json"})
		// With([]string{""npm", "sbom", "--sbom-type", "application", "--sbom-format", "spdx", ">", "sbom.spdx.json"}).Terminal()

	builder := installer.WithEnvVariable("NODE_ENV", "production").WithExec([]string{"npm", "run", "build"})

	outputs = outputs.WithFile("dist/main.js", builder.File("dist/main.js")).
		WithFile("dist/styles.css", builder.File("dist/styles.css")).
		WithFile("dist/manifest.json", builder.File("manifest.json")).
		WithFile("dist/sbom.spdx.json", installer.File("sbom.spdx.json"))

	_, err := outputs.Export(ctx, ".")
	if err != nil {
		return err
	}
	return nil
}
