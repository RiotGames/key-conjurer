module github.com/riotgames/key-conjurer/cli

require (
	github.com/go-ini/ini v1.61.0
	github.com/hashicorp/go-rootcerts v1.0.2
	github.com/mattn/go-runewidth v0.0.9 // indirect
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mitchellh/go-ps v1.0.0
	github.com/olekukonko/tablewriter v0.0.4
	github.com/riotgames/key-conjurer/api v0.0.0-20200910171920-2c564e9fc301
	github.com/smartystreets/goconvey v0.0.0-20190330032615-68dc04aab96a // indirect
	github.com/spf13/cobra v1.0.0
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/stretchr/testify v1.6.1
	golang.org/x/crypto v0.0.0-20201016220609-9e8e0b390897
	golang.org/x/sys v0.0.0-20200909081042-eff7692f9009 // indirect
	gopkg.in/ini.v1 v1.42.0 // indirect
)

go 1.13

replace github.com/riotgames/key-conjurer/api => ../api
