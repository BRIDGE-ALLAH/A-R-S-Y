package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"github.com/bridge-allah/a-r-s-y"
	"github.com/bridge-allah/a-r-s-y/apis"
	"github.com/bridge-allah/a-r-s-y/arsy/config"
	"github.com/bridge-allah/a-r-s-y/arsy/httpapi"
	_ "github.com/bridge-allah/a-r-s-y/arsy/migrations"
	"github.com/bridge-allah/a-r-s-y/core"
	"github.com/bridge-allah/a-r-s-y/plugins/jsvm"
	"github.com/bridge-allah/a-r-s-y/plugins/migratecmd"
	"github.com/bridge-allah/a-r-s-y/tools/hook"
	"github.com/bridge-allah/a-r-s-y/tools/osutils"

	"github.com/spf13/cobra"
)

func main() {
	app := pocketbase.New()

	arsyConfig, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	httpapi.Register(app, arsyConfig, pocketbase.Version)

	app.OnServe().Bind(&hook.Handler[*core.ServeEvent]{
		Id: "arsy_data_directories",
		Func: func(e *core.ServeEvent) error {
			if _, err := config.EnsureDirectories(arsyConfig.DataDir); err != nil {
				return err
			}
			return e.Next()
		},
	})

	app.RootCmd.AddCommand(&cobra.Command{
		Use:   "doctor",
		Short: "check the local A-R-S-Y control-plane setup",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("Product: A-R-S-Y %s\n", pocketbase.Version)
			fmt.Printf("Go runtime: %s\n", runtime.Version())
			fmt.Printf("Data directory: %s\n", arsyConfig.DataDir)
			dirs, err := config.EnsureDirectories(arsyConfig.DataDir)
			if err != nil {
				fmt.Printf("Data directory: error (%v)\n", err)
				return err
			}
			fmt.Printf("Data directories: accessible (%s, %s)\n", dirs.Temp, dirs.Manifests)
			secretConfigured := os.Getenv(arsyConfig.SecretKeyEnv) != ""
			if !secretConfigured {
				fmt.Printf("Secret key: not configured (set %s)\n", arsyConfig.SecretKeyEnv)
			} else {
				fmt.Println("Secret key: configured (value hidden)")
			}
			if !app.IsBootstrapped() {
				fmt.Println("Database: not ready")
				return fmt.Errorf("A-R-S-Y database is not ready")
			}
			if _, err := app.DB().NewQuery("SELECT 1").Execute(); err != nil {
				fmt.Printf("Database: unavailable (%v)\n", err)
				return err
			}
			fmt.Println("Database: ready")
			if !secretConfigured {
				return fmt.Errorf("A-R-S-Y secret key is not configured")
			}
			return nil
		},
	})

	// ---------------------------------------------------------------
	// Optional plugin flags:
	// ---------------------------------------------------------------

	var hooksDir string
	app.RootCmd.PersistentFlags().StringVar(
		&hooksDir,
		"hooksDir",
		"",
		"the directory with the JS app hooks",
	)

	var hooksWatch bool
	app.RootCmd.PersistentFlags().BoolVar(
		&hooksWatch,
		"hooksWatch",
		true,
		"auto restart the app on pb_hooks file change; it has no effect on Windows",
	)

	var hooksPool int
	app.RootCmd.PersistentFlags().IntVar(
		&hooksPool,
		"hooksPool",
		15,
		"the total prewarm goja.Runtime instances for the JS app hooks execution",
	)

	var migrationsDir string
	app.RootCmd.PersistentFlags().StringVar(
		&migrationsDir,
		"migrationsDir",
		"",
		"the directory with the user defined migrations",
	)

	var automigrate bool
	app.RootCmd.PersistentFlags().BoolVar(
		&automigrate,
		"automigrate",
		true,
		"enable/disable auto migrations",
	)

	var publicDir string
	app.RootCmd.PersistentFlags().StringVar(
		&publicDir,
		"publicDir",
		defaultPublicDir(),
		"the directory to serve static files",
	)

	var indexFallback bool
	app.RootCmd.PersistentFlags().BoolVar(
		&indexFallback,
		"indexFallback",
		true,
		"fallback the request to index.html on missing static path, e.g. when pretty urls are used with SPA",
	)

	app.RootCmd.ParseFlags(os.Args[1:])

	// ---------------------------------------------------------------
	// Plugins and hooks:
	// ---------------------------------------------------------------

	// load jsvm (pb_hooks and pb_migrations)
	jsvm.MustRegister(app, jsvm.Config{
		MigrationsDir: migrationsDir,
		HooksDir:      hooksDir,
		HooksWatch:    hooksWatch,
		HooksPoolSize: hooksPool,
	})

	// migrate command (with js templates)
	migratecmd.MustRegister(app, app.RootCmd, migratecmd.Config{
		TemplateLang: migratecmd.TemplateLangJS,
		Automigrate:  automigrate,
		Dir:          migrationsDir,
	})


	// static route to serves files from the provided public dir
	// (if publicDir exists and the route path is not already defined)
	app.OnServe().Bind(&hook.Handler[*core.ServeEvent]{
		Func: func(e *core.ServeEvent) error {
			if !e.Router.HasRoute(http.MethodGet, "/{path...}") {
				e.Router.GET("/{path...}", apis.Static(os.DirFS(publicDir), indexFallback))
			}

			return e.Next()
		},
		Priority: 999, // execute as latest as possible to allow users to provide their own route
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}

// the default pb_public dir location is relative to the executable
func defaultPublicDir() string {
	if osutils.IsProbablyGoRun() {
		return "./pb_public"
	}

	return filepath.Join(os.Args[0], "../pb_public")
}
