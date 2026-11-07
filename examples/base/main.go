package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/bridge-allah/a-r-s-y"
	"github.com/bridge-allah/a-r-s-y/apis"
	arsyapp "github.com/bridge-allah/a-r-s-y/arsy"
	arsyconfig "github.com/bridge-allah/a-r-s-y/arsy/config"
	"github.com/bridge-allah/a-r-s-y/arsy/diagnostics"
	"github.com/bridge-allah/a-r-s-y/arsy/httpapi"
	_ "github.com/bridge-allah/a-r-s-y/arsy/migrations"
	arsysecurity "github.com/bridge-allah/a-r-s-y/arsy/security"
	"github.com/bridge-allah/a-r-s-y/core"
	"github.com/bridge-allah/a-r-s-y/plugins/jsvm"
	"github.com/bridge-allah/a-r-s-y/plugins/migratecmd"
	"github.com/bridge-allah/a-r-s-y/tools/hook"
	"github.com/bridge-allah/a-r-s-y/tools/osutils"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "generate-secret-key" {
		command := arsysecurity.NewGenerateKeyCommand(os.Stdout)
		command.SetArgs(os.Args[2:])
		if err := command.Execute(); err != nil {
			log.Fatal(err)
		}
		return
	}

	runtimeConfig, err := arsyconfig.Load()
	if err != nil {
		log.Fatalf("invalid A-R-S-Y configuration: %v", err)
	}
	app := pocketbase.NewWithConfig(pocketbase.Config{
		DefaultDev:     osutils.IsProbablyGoRun(),
		DefaultDataDir: runtimeConfig.DataDir,
	})
	app.OnBootstrap().BindFunc(func(e *core.BootstrapEvent) error {
		if err := e.Next(); err != nil {
			return err
		}
		_, err := arsyapp.EnsureDataLayout(e.App.DataDir())
		return err
	})
	httpapi.RegisterRoutes(app, runtimeConfig.APIPrefix, pocketbase.Version, runtimeConfig.SecretKeyEnv)
	httpapi.RegisterConsole(app, runtimeConfig.ConsoleRoute)

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

	app.RootCmd.AddCommand(diagnostics.NewDoctorCommand(app, pocketbase.Version, runtimeConfig.SecretKeyEnv))
	app.RootCmd.AddCommand(arsysecurity.NewGenerateKeyCommand(os.Stdout))

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
