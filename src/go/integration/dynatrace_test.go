package integration_test

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/libbuildpack/cutlass"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testDynatrace(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		app      *cutlass.App
		services []string
	)

	it.Before(func() {
		app = cutlass.New(filepath.Join(settings.FixturesPath, "default", "simple"))
		app.SetEnv("BP_DEBUG", "true")
		PushAppAndConfirm(t, app)
	})

	it.After(func() {
		app = DestroyApp(t, app)

		for _, service := range services {
			command := exec.Command("cf", "delete-service", "-f", service)
			_, err := command.Output()
			Expect(err).NotTo(HaveOccurred())
		}
	})

	context("when deploying with Dynatrace agent with single credentials service", func() {
		it("checks if Dynatrace injection was successful", func() {
			service := "dynatrace-" + cutlass.RandStringRunes(20) + "-service"
			command := exec.Command("cf", "cups", service, "-p", fmt.Sprintf("'{\"apitoken\":\"secretpaastoken\",\"apiurl\":\"%s\",\"environmentid\":\"envid\"}'", settings.Dynatrace.URI))
			output, err := command.CombinedOutput()
			Expect(err).NotTo(HaveOccurred(), string(output))

			services = append(services, service)

			command = exec.Command("cf", "bind-service", app.Name, service)
			output, err = command.CombinedOutput()
			Expect(err).NotTo(HaveOccurred(), string(output))

			command = exec.Command("cf", "restage", app.Name)
			output, err = command.CombinedOutput()
			Expect(err).NotTo(HaveOccurred(), string(output))

			Expect(app.ConfirmBuildpack(settings.Buildpack.Version)).To(Succeed())
			Expect(app.Stdout.String()).To(ContainSubstring("Dynatrace service credentials found. Setting up Dynatrace OneAgent."))
			Expect(app.Stdout.String()).To(ContainSubstring("Starting Dynatrace OneAgent installer"))
			Expect(app.Stdout.String()).To(ContainSubstring("Copy dynatrace-env.sh"))
			Expect(app.Stdout.String()).To(ContainSubstring("Dynatrace OneAgent installed."))
			Expect(app.Stdout.String()).To(ContainSubstring("Dynatrace OneAgent injection is set up."))
		})
	})

	context("deploying a Go app with Dynatrace agent with configured network zone", func() {
		it("checks if networkzone setting was successful", func() {
			serviceName := "dynatrace-" + cutlass.RandStringRunes(20) + "-service"
			command := exec.Command("cf", "cups", serviceName, "-p", fmt.Sprintf("'{\"apitoken\":\"secretpaastoken\",\"apiurl\":\"%s\",\"environmentid\":\"envid\", \"networkzone\":\"testzone\"}'", settings.Dynatrace.URI))
			_, err := command.CombinedOutput()
			Expect(err).To(BeNil())
			services = append(services, serviceName)

			command = exec.Command("cf", "bind-service", app.Name, serviceName)
			_, err = command.CombinedOutput()
			Expect(err).To(BeNil())
			command = exec.Command("cf", "restage", app.Name)
			_, err = command.Output()
			Expect(err).To(BeNil())

			Expect(app.ConfirmBuildpack(settings.Buildpack.Version)).To(Succeed())
			Expect(app.Stdout.String()).To(ContainSubstring("Dynatrace service credentials found. Setting up Dynatrace OneAgent."))
			Expect(app.Stdout.String()).To(ContainSubstring("Starting Dynatrace OneAgent installer"))
			Expect(app.Stdout.String()).To(ContainSubstring("Copy dynatrace-env.sh"))
			Expect(app.Stdout.String()).To(ContainSubstring("Setting DT_NETWORK_ZONE..."))
			Expect(app.Stdout.String()).To(ContainSubstring("Dynatrace OneAgent installed."))
			Expect(app.Stdout.String()).To(ContainSubstring("Dynatrace OneAgent injection is set up."))
		})
	})

	context("when deploying with Dynatrace agent with two credentials services", func() {
		it("checks if detection of second service with credentials works", func() {
			service := "dynatrace-" + cutlass.RandStringRunes(20) + "-service"
			command := exec.Command("cf", "cups", service, "-p", fmt.Sprintf("'{\"apitoken\":\"secretpaastoken\",\"apiurl\":\"%s\",\"environmentid\":\"envid\"}'", settings.Dynatrace.URI))
			output, err := command.CombinedOutput()
			Expect(err).NotTo(HaveOccurred(), string(output))

			services = append(services, service)

			command = exec.Command("cf", "bind-service", app.Name, service)
			output, err = command.CombinedOutput()
			Expect(err).NotTo(HaveOccurred(), string(output))

			service = "dynatrace-dupe-" + cutlass.RandStringRunes(20) + "-service"
			command = exec.Command("cf", "cups", service, "-p", fmt.Sprintf("'{\"apitoken\":\"secretpaastoken\",\"apiurl\":\"%s\",\"environmentid\":\"envid\"}'", settings.Dynatrace.URI))
			output, err = command.CombinedOutput()
			Expect(err).NotTo(HaveOccurred(), string(output))

			services = append(services, service)

			command = exec.Command("cf", "bind-service", app.Name, service)
			output, err = command.CombinedOutput()
			Expect(err).NotTo(HaveOccurred(), string(output))

			command = exec.Command("cf", "restage", app.Name)
			output, err = command.CombinedOutput()
			Expect(err).NotTo(HaveOccurred(), string(output))

			Expect(app.Stdout.String()).To(ContainSubstring("More than one matching service found!"))
		})
	})

	context("when deploying with Dynatrace agent with failing agent download and ignoring errors", func() {
		it("checks if skipping download errors works", func() {
			service := "dynatrace-" + cutlass.RandStringRunes(20) + "-service"
			command := exec.Command("cf", "cups", service, "-p", fmt.Sprintf("'{\"apitoken\":\"secretpaastoken\",\"apiurl\":\"%s/no-such-endpoint\",\"environmentid\":\"envid\",\"skiperrors\":\"true\"}'", settings.Dynatrace.URI))
			output, err := command.CombinedOutput()
			Expect(err).NotTo(HaveOccurred(), string(output))

			services = append(services, service)

			command = exec.Command("cf", "bind-service", app.Name, service)
			output, err = command.CombinedOutput()
			Expect(err).NotTo(HaveOccurred(), string(output))

			command = exec.Command("cf", "restage", app.Name)
			output, err = command.CombinedOutput()
			Expect(err).NotTo(HaveOccurred(), string(output))

			Expect(app.Stdout.String()).To(ContainSubstring("Download returned with status 404"))
			Expect(app.Stdout.String()).To(ContainSubstring("Error during installer download, skipping installation"))
		})
	})

	context("deploying a with Dynatrace agent with two dynatrace services", func() {
		it("check if service detection isn't disturbed by a service with tags", func() {
			service := "dynatrace-" + cutlass.RandStringRunes(20) + "-service"
			command := exec.Command("cf", "cups", service, "-p", fmt.Sprintf("'{\"apitoken\":\"secretpaastoken\",\"apiurl\":\"%s\",\"environmentid\":\"envid\"}'", settings.Dynatrace.URI))
			output, err := command.CombinedOutput()
			Expect(err).NotTo(HaveOccurred(), string(output))

			services = append(services, service)

			command = exec.Command("cf", "bind-service", app.Name, service)
			output, err = command.CombinedOutput()
			Expect(err).NotTo(HaveOccurred(), string(output))

			service = "dynatrace-tags-" + cutlass.RandStringRunes(20) + "-service"
			command = exec.Command("cf", "cups", service, "-p", "'{\"tag:dttest\":\"dynatrace_test\"}'")
			output, err = command.CombinedOutput()
			Expect(err).NotTo(HaveOccurred(), string(output))

			services = append(services, service)

			command = exec.Command("cf", "bind-service", app.Name, service)
			output, err = command.CombinedOutput()
			Expect(err).NotTo(HaveOccurred(), string(output))

			command = exec.Command("cf", "restage", app.Name)
			output, err = command.CombinedOutput()
			Expect(err).NotTo(HaveOccurred(), string(output))

			Expect(app.ConfirmBuildpack(settings.Buildpack.Version)).To(Succeed())
			Expect(app.Stdout.String()).To(ContainSubstring("Dynatrace service credentials found. Setting up Dynatrace OneAgent."))
			Expect(app.Stdout.String()).To(ContainSubstring("Starting Dynatrace OneAgent installer"))
			Expect(app.Stdout.String()).To(ContainSubstring("Copy dynatrace-env.sh"))
			Expect(app.Stdout.String()).To(ContainSubstring("Dynatrace OneAgent installed."))
			Expect(app.Stdout.String()).To(ContainSubstring("Dynatrace OneAgent injection is set up."))
		})
	})

	context("deploying with Dynatrace agent with single credentials service and without manifest.json", func() {
		it("checks if Dynatrace injection was successful", func() {
			service := "dynatrace-" + cutlass.RandStringRunes(20) + "-service"
			command := exec.Command("cf", "cups", service, "-p", fmt.Sprintf("'{\"apitoken\":\"secretpaastoken\",\"apiurl\":\"%s\",\"environmentid\":\"envid\"}'", settings.Dynatrace.URI))
			output, err := command.CombinedOutput()
			Expect(err).NotTo(HaveOccurred(), string(output))

			services = append(services, service)

			command = exec.Command("cf", "bind-service", app.Name, service)
			output, err = command.CombinedOutput()
			Expect(err).NotTo(HaveOccurred(), string(output))

			command = exec.Command("cf", "restage", app.Name)
			output, err = command.CombinedOutput()
			Expect(err).NotTo(HaveOccurred(), string(output))

			Expect(app.ConfirmBuildpack(settings.Buildpack.Version)).To(Succeed())
			Expect(app.Stdout.String()).To(ContainSubstring("Dynatrace service credentials found. Setting up Dynatrace OneAgent."))
			Expect(app.Stdout.String()).To(ContainSubstring("Starting Dynatrace OneAgent installer"))
			Expect(app.Stdout.String()).To(ContainSubstring("Copy dynatrace-env.sh"))
			Expect(app.Stdout.String()).To(ContainSubstring("Dynatrace OneAgent installed."))
			Expect(app.Stdout.String()).To(ContainSubstring("Dynatrace OneAgent injection is set up."))
		})
	})

	context("deployingwith Dynatrace agent with failing agent download and checking retry", func() {
		it("checks if retrying downloads works", func() {
			service := "dynatrace-" + cutlass.RandStringRunes(20) + "-service"
			command := exec.Command("cf", "cups", service, "-p", fmt.Sprintf("'{\"apitoken\":\"secretpaastoken\",\"apiurl\":\"%s/no-such-endpoint\",\"environmentid\":\"envid\"}'", settings.Dynatrace.URI))
			output, err := command.CombinedOutput()
			Expect(err).NotTo(HaveOccurred(), string(output))

			services = append(services, service)

			command = exec.Command("cf", "bind-service", app.Name, service)
			output, err = command.CombinedOutput()
			Expect(err).NotTo(HaveOccurred(), string(output))

			command = exec.Command("cf", "restage", app.Name)
			output, err = command.CombinedOutput()
			Expect(err).To(HaveOccurred(), string(output))
			Expect(app.Stdout.String()).To(ContainSubstring("Failed to compile droplet"))

			Expect(app.Stdout.String()).To(ContainSubstring("Error during installer download, retrying in 4s"))
			Expect(app.Stdout.String()).To(ContainSubstring("Error during installer download, retrying in 5s"))
			Expect(app.Stdout.String()).To(ContainSubstring("Error during installer download, retrying in 7s"))
			Expect(app.Stdout.String()).To(ContainSubstring("Download returned with status 404"))
		})
	})

	context("deploying with Dynatrace agent with single credentials service and a redis service", func() {
		it("checks if Dynatrace injection was successful", func() {
			service := "dynatrace-" + cutlass.RandStringRunes(20) + "-service"
			command := exec.Command("cf", "cups", service, "-p", fmt.Sprintf("'{\"apitoken\":\"secretpaastoken\",\"apiurl\":\"%s\",\"environmentid\":\"envid\"}'", settings.Dynatrace.URI))
			output, err := command.CombinedOutput()
			Expect(err).NotTo(HaveOccurred(), string(output))

			services = append(services, service)

			command = exec.Command("cf", "bind-service", app.Name, service)
			output, err = command.CombinedOutput()
			Expect(err).NotTo(HaveOccurred(), string(output))

			service = "redis-" + cutlass.RandStringRunes(20) + "-service"
			command = exec.Command("cf", "cups", service, "-p", "'{\"name\":\"redis\", \"credentials\":{\"db_type\":\"redis\", \"instance_administration_api\":{\"deployment_id\":\"12345asdf\", \"instance_id\":\"12345asdf\", \"root\":\"https://doesnotexi.st\"}}}'")
			output, err = command.CombinedOutput()
			Expect(err).NotTo(HaveOccurred(), string(output))

			services = append(services, service)

			command = exec.Command("cf", "bind-service", app.Name, service)
			output, err = command.CombinedOutput()
			Expect(err).NotTo(HaveOccurred(), string(output))

			command = exec.Command("cf", "restage", app.Name)
			output, err = command.CombinedOutput()
			Expect(err).NotTo(HaveOccurred(), string(output))

			Expect(app.ConfirmBuildpack(settings.Buildpack.Version)).To(Succeed())
			Expect(app.Stdout.String()).To(ContainSubstring("Dynatrace service credentials found. Setting up Dynatrace OneAgent."))
			Expect(app.Stdout.String()).To(ContainSubstring("Starting Dynatrace OneAgent installer"))
			Expect(app.Stdout.String()).To(ContainSubstring("Copy dynatrace-env.sh"))
			Expect(app.Stdout.String()).To(ContainSubstring("Dynatrace OneAgent installed."))
			Expect(app.Stdout.String()).To(ContainSubstring("Dynatrace OneAgent injection is set up."))
		})
	})
}
