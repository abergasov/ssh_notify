# Info

Notify about new ssh login on your server into telegram/slack

## Sample
> notify on key login

![Repo_List](log_by_key.png)


> notify on password login

![Repo_List](log_by_pass.png)

> notify if list of known ips specified (known/unknown ip login)

![Repo_List](ip_spec.png)

## Build
You'll need go v1.13 or later

### Install Go
```shell script
sudo add-apt-repository ppa:longsleep/golang-backports
sudo apt-get update
sudo apt-get install golang-go
# or via snap
snap install go --classic
```

### Ansible script
script for create for `admin` user
roles/ssh_notify/tasks/main.yml
```yaml
---
- name: Install go
  community.general.snap:
    name: go
    classic: yes
    channel: 1.18/stable

- name: Clone a github repository
  become: yes
 # become_user: admin
  become_method: sudo
  git:
    repo: https://github.com/abergasov/ssh_notify
    dest: /home/admin/go/src/ssh_notify
    clone: yes
    update: yes

- name: Run install target
  timeout: 30
  no_log: True
  become: yes
  become_method: sudo
  ansible.builtin.command: make name={{ server_name }} tg_token={{ tg_token }} tg_chat={{ tg_chat }} install
  args:
    chdir: /home/admin/go/src/ssh_notify
```

playbook.yaml
```yaml
# VPN
---
- hosts: "vpns"
  user: root
  vars_files:
    - secret
    - vars/main.yml
  become: true
  roles:
    ...
    - ssh_notify
    ...
```

### Install via make
```shell script
mkdir -p "$HOME/go/src"
cd "$HOME/go/src"
git clone https://github.com/abergasov/ssh_notify.git
```
Create with telegram notify only
```shell script
 make name=dev_server tg_token=12312312312 tg_chat=-123 install
```

Create with full config
```shell script
 make name=dev_server tg_token=123122 tg_chat=-123 sl_token=123 sl_chn=123 ips=127.0.0.1:PersonalVPN install
```

Remove
```shell script
make remove
```

### Manually install
```shell script
mkdir -p "$HOME/go/src"
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

# list of known ips separated by comma (deploy bot, personal vpn, etc...)
KnownIps = 35.243.248.170:Gitlab ; 35.190.190.84 :Gitlab ; 35.229.20.217:Gitlab;127.0.0.1:PersonalVPN
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
