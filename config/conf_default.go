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

	return conf
}
