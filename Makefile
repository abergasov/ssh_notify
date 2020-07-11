COMMIT?=$(shell git ls-files | xargs sha256sum | cut -d" " -f1 | sha256sum | cut -d" " -f1)
BUILD_TIME?=$(shell date -u '+%Y-%m-%d_%H:%M:%S')

TG_TOKEN := $(or ${tg_token},${tg_token},YOUR_TELEGRAM_BOT_TOKEN_HERE)
TG_CHAT := $(or ${tg_chat},${tg_chat},YOUR_TELEGRAM_CHAT_HERE)
SLACK_TOKEN := $(or ${sl_token},${sl_token},YOUR_SLACK_BOT_TOKEN_HERE)
SLACK_CHANEL := $(or ${sl_chn},${sl_chn},YOUR_SLACK_CHANNEL_HERE)
KNOWN_IPS := $(or ${ips},${ips},SET_KNOWN_IP_LIST_HERE)

install:
	@echo "-- current dir: ${CURRENT_DIR}"

	@echo "-- building binary"
	go build -ldflags "-X main.buildHash=${COMMIT} -X main.buildTime=${BUILD_TIME}" -o ./bin
	@echo "-- copy binary"
	cp ./bin/ssh_notify /usr/bin/

	@echo "-- copy sample config"
	sudo cp ssh_notify.conf /etc/ssh_notify.conf

	@echo "-- set config"

#	@echo "${TG_TOKEN}"
#	@echo "${SLACK_TOKEN}"
#	@echo "${TG_CHAT}"

	@echo "-- creating service"
	sudo mkdir -p /etc/systemd/system
	sudo cp ssh_notify.service /etc/systemd/system/ssh_notify.service

	@echo "-- enable service"
	sudo service ssh_notify start && sudo systemctl enable ssh_notify

remove:
	@echo "-- remove service"
	sudo service ssh_notify stop
	sudo systemctl disable ssh_notify
	sudo rm -rfv /etc/systemd/system/ssh_notify.service