# simple-backup
A simple backup program in golang. Uses cp -l (copy with hardlinks) and rsync to
create incremental backups to an remote server via ssh

## How to install

First you must install build enviroment

### Mac OS 

<pre>
$ brew install go
$ brew install dep
</pre>

### Linux

<pre>
$ yum install golang

... or ...

$ apt-get install golang

$ wget https://github.com/golang/dep/releases/download/v0.3.2/dep-linux-amd64
$ sudo mv dep-linux-amd64 /usr/local/bin/dep
</pre>


### Build

<pre>
$ git clone https://github.com/johandalabacka/simple-backup
$ cd simple-backup
$ dep ensure

$ go build -o simple-backup *.go
</pre>

### Install

<pre>
sudo cp simple-backup /usr/local/bin
sudo cp simple-backup-example.toml /etc/simple-backup.toml
</pre>

Edit /etc/simple-backup.toml to match your setup