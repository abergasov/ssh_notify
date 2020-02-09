# Info

Notify about new ssh login on your server into telegram/slack

## Sample
> notify on key login

![Repo_List](log_by_key.png)


> notify on password login

![Repo_List](log_by_pass.png)

## Build
You'll need go v1.13 or later

### Install Go
```shell script
sudo apt update && sudo apt install golang-go
```

### Clone repository and build
```shell script
mkdir "$HOME/go/src"
cd "$HOME/go/src"
git clone https://github.com/abergasov/ssh_notify.git
cd ssh_notify
go build main.go
```

### Set config
```shell script
sudo touch /etc/ssh_notify.conf && sudo nano /etc/ssh_notify.conf 
```

Sample config 
```shell script                                                                   /etc/ssh_notify.conf                                                                               
SSHLogFile = /var/log/auth.log
ServerName = test_test_server_name

# telegram settings (optional, just do not set if not need)
TelegramBotToken = YOUR_BOT_TOKEN_HERE
TelegramNotifyChat = YOUR_CHAT_HERE

# slack settings (optional, just do not set if not need)
SlackBotToken = YOUR_BOT_TOKEN_HERE
SlackTargetChannel = YOUR_CHANNEL_HERE
```

### Create service and run
```shell script
sudo nano /lib/systemd/system/ssh_notify.service
```
Put text
```shell script
[Unit]
Description=notify on every ssh_login

[Service]
Type=simple
Restart=always
RestartSec=5s
ExecStart=PATH_TO_HOMEDIR/go/src/ssh_notify/main

[Install]
WantedBy=multi-user.target
```

Start service
```shell script
sudo service ssh_notify start
sudo systemctl enable ssh_notify
```

### Logs
```bash
sudo journalctl -f -u ssh_notify.service
```