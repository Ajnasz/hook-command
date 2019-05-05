module github.com/Ajnasz/hook-command

replace github.com/Sirupsen/logrus => github.com/sirupsen/logrus v1.4.1

go 1.12

require (
	github.com/Ajnasz/logrus-redis v0.0.0-20180406191141-0cf75d7e4a80
	github.com/Sirupsen/logrus v0.0.0-00010101000000-000000000000
	github.com/coreos/go-systemd v0.0.0-20190321100706-95778dfbb74e
	github.com/go-redis/redis v6.15.2+incompatible
	github.com/kelseyhightower/envconfig v1.3.0
)
