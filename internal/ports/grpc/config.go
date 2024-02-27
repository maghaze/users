package grpc

type Config struct {
	Targets struct {
		Auth string `koanf:"auth"`
	} `koanf:"targets"`
}
