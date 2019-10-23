#!/bin/sh
add-apt-repository ppa:longsleep/golang-backports
apt-get update
apt-get install golang-go git -y
git clone --recursive git@github.com:xxzl0130/AntiRivercrab.git
cd AntiRivercrab
git checkout server
go build -o AntiRivercrab ./main.go
mkdir /etc/AntiRivercrab
cp ./AntiRivercrab /usr/bin/
cp -r ./PACFile /etc/AntiRivercrab/
cp ./ar.sh /etc/init.d
chmod 755 /etc/init.d/ar.sh
update-rc.d ar.sh defaults 95