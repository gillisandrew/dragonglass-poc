package main

import (
	"context"
	"fmt"
	"os"

	"dagger.io/dagger/dag"
)

func main() {
	if len(os.Args) < 4 {
		fmt.Fprintf(os.Stderr, "Usage: %s <repository> <ref> <plugin>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Example: %s https://github.com/gillisandrew/dragonglass-poc.git main example-plugin\n", os.Args[0])
		os.Exit(1)
	}

	repository := os.Args[1]
	ref := os.Args[2]
	plugin := os.Args[3]

	if err := build(context.Background(), repository, ref, plugin); err != nil {
		fmt.Println(err)
	}
}

func build(ctx context.Context, repository, ref, plugin string) error {
	fmt.Println("Building with Dagger")
	defer dag.Close()

	// create empty directory to put build outputs
	outputs := dag.Directory()
	repo := dag.Git(repository).Ref(ref).Tree()
	installer := dag.Container().
		From("node:22").
		WithDirectory("/usr/src/plugin", repo.Directory(plugin)).
		WithWorkdir("/usr/src/plugin").
		WithExec([]string{"npm", "ci"}).
		WithExec([]string{"bash", "-c", "npm sbom --sbom-type application --sbom-format spdx > sbom.spdx.json"})
		// With([]string{""npm", "sbom", "--sbom-type", "application", "--sbom-format", "spdx", ">", "sbom.spdx.json"}).Terminal()

	builder := installer.WithEnvVariable("NODE_ENV", "production").WithExec([]string{"npm", "run", "build"})

	outputs = outputs.WithFile("main.js", builder.File("dist/main.js")).
		WithFile("styles.css", builder.File("dist/styles.css")).
		WithFile("manifest.json", builder.File("manifest.json")).
		WithFile("sbom.spdx.json", installer.File("sbom.spdx.json"))

	_, err := outputs.Export(ctx, "dist")
	if err != nil {
		return err
	}
	return nil
}
