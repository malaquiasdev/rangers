package learn

import "context"

func Command(ctx context.Context, args []string, env func(string) string, dep *Dependencies) error {
	metrics := len(args) > 0 && args[0] == "metrics"
	flags := args
	if metrics {
		flags = args[1:]
	}
	cfg, err := ReadConfig(flags, env)
	if err != nil {
		return err
	}
	if metrics {
		return Metrics(&cfg, dep.Out)
	}
	return Run(ctx, &cfg, dep)
}
