#!/bin/bash
set -eu

RELEASE=v1.0.1

case "$(uname -m)" in
	x86_64 ) ARCH="amd64";;
	i386 ) ARCH="386";;
	* ) echo "Your platform ($(uname -a)) is not supported.";exit 1;;
esac

echo -e "\ndownloading mydnsnotifier..."

wget "https://github.com/hrntknr/mydns-notifier/releases/download/${RELEASE}/linux_${ARCH}_mydnsnotifier" -O /usr/local/bin/mydnsnotifier -q
chmod +x /usr/local/bin/mydnsnotifier

read -p "mydns id: " id
read -p "mydns password: " password

cat << EOS > /etc/mydnsnotifier.toml
[notice]
id = "${id}"
password = "${password}"
cron = "@hourly"
EOS

while true; do
	read -p 'enable slack? [Y/n]: ' answer
	case $answer in
		[Yy]* )
			slack=true
			break
			;;
		[Nn]* )
			slack=false
			break
			;;
		* )
			echo Please answer Y/N.
	esac
done;

if $slack;then
	read -p "slack webhook url: " hookURL
	cat << EOS >> /etc/mydnsnotifier.toml
[log.slack]
enable = true
hookURL = "${hookURL}"
EOS
fi

cat << EOS > /etc/systemd/system/mydnsnotifier.service
[Unit]
Description=mydnsnotifier service
After=network-online.target

[Service]
Type=simple
ExecStart=/usr/local/bin/mydnsnotifier --config /etc/mydnsnotifier.toml
Restart=always

[Install]
WantedBy=network-online.target
EOS

systemctl daemon-reload
systemctl start mydnsnotifier
systemctl enable mydnsnotifier

echo "Setting complete."
