#!/bin/sh
sudo su
apt-get update
apt-get install golang-go git -y
git clone git@github.com:xxzl0130/AntiRivercrab.git
cd AntiRivercrab
git checkout server
go build -o AntiRivercrab ./main.go
mkdir /usr/bin/AntiRivercrab
mkdir /etc/AntiRivercrab
cp ./AntiRivercrab /usr/bin/AntiRivercrab/
cp -r ./PACFile /etc/AntiRivercrab/
cp ./ar.sh /etc/init.d
chmod 755 /etc/init.d/ar.sh
update-rc.d /etc/init.d/ar.sh defaults 95