# Go Repository Template

A reusable starting point for Go repositories. Create a new GitHub repository from this template to get a consistent project foundation without copying project-specific implementation code.

The template includes:

- a Makefile for formatting, linting, tests, vulnerability checks, and dependency maintenance;
- pinned Go tooling versions in `.versions`;
- `golangci-lint` configuration and a local tool cache;
- continuous integration and release workflows;
- Dependabot updates for Go modules and GitHub Actions;
- a workflow for keeping the pinned linting version current;
- semantic versioning support through pull requests and releases.

After creating a repository from this template, update the module path in `go.mod`, rename the project references in the documentation, and replace the example command with the project’s own implementation. The shared tooling can then remain unchanged as the project evolves.
