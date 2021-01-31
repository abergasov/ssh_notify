COMMIT?=$(shell git ls-files | xargs sha256sum | cut -d" " -f1 | sha256sum | cut -d" " -f1)
BUILD_TIME?=$(shell date -u '+%Y-%m-%d_%H:%M:%S')

SERVER_NAME := $(or ${name},${name},"")
TG_TOKEN := $(or ${tg_token},${tg_token},"")
TG_CHAT := $(or ${tg_chat},${tg_chat},"")
SLACK_TOKEN := $(or ${sl_token},${sl_token},"")
SLACK_CHANEL := $(or ${sl_chn},${sl_chn},"")
KNOWN_IPS := $(or ${ips},${ips},"")

install:
	@echo "-- building binary"
	go build -ldflags "-X main.buildHash=${COMMIT} -X main.buildTime=${BUILD_TIME}" -o ./bin/ssh_notify ./cmd
	@echo "-- copy binary"
	#sudo cp ./bin/ssh_notify /usr/bin/

	@echo "-- create sample config"
	cp ssh_notify.conf ssh_notify.conf.tmp

	@echo "-- set config values"

	@sed -i s/test_test_server_name/$(SERVER_NAME)/g ssh_notify.conf.tmp
	@sed -i s/YOUR_TELEGRAM_BOT_TOKEN_HERE/$(TG_TOKEN)/g ssh_notify.conf.tmp
	@sed -i s/YOUR_TELEGRAM_CHAT_HERE/$(TG_CHAT)/g ssh_notify.conf.tmp
	@sed -i s/YOUR_SLACK_BOT_TOKEN_HERE/$(SLACK_TOKEN)/g ssh_notify.conf.tmp
	@sed -i s/YOUR_SLACK_CHANNEL_HERE/$(SLACK_CHANEL)/g ssh_notify.conf.tmp
	@sed -i s/SET_KNOWN_IP_LIST_HERE/$(KNOWN_IPS)/g ssh_notify.conf.tmp

	sudo cp ssh_notify.conf.tmp /etc/ssh_notify.conf
	rm ssh_notify.conf.tmp

	@echo "-- creating service"
	sudo mkdir -p /etc/systemd/system
	sudo cp ssh_notify.service /etc/systemd/system/ssh_notify.service

	@echo "-- enable service"
	sudo service ssh_notify start && sudo systemctl enable ssh_notify

remove:
	@echo "-- remove service"
	sudo service ssh_notify stop
	sudo systemctl disable ssh_notify
	sudo rm /etc/systemd/system/ssh_notify.service
	sudo rm /etc/ssh_notify.conf