package config

func Default() *Conf {
	conf := &Conf{}
	conf.Storage.SrvPath = "./etc/srv"
	conf.Storage.SrvDbName = "service.db"
	conf.Storage.UsrPath = "./etc/usr"

	conf.Log.Stdio    = true
	conf.Log.File     = true
	conf.Log.FilePath = "./etc/log"
	conf.Log.Syslog   = false
	conf.Log.Systag   = "srv"

	conf.User.Root = "root"
	conf.User.Default ="default"

	conf.Network.Http.Enable = true
	conf.Network.Http.Listen = "0.0.0.0"
	conf.Network.Http.Port = "8080"
	conf.Network.Http.ReadTimeout = "15s"
	conf.Network.Http.WriteTimeout = "15s"

	return conf
}
