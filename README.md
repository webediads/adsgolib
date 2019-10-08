Modifier ~/.ssh/config
```
Host git.webedia-group.net
    User git
    IdentityFile ~/.ssh/id_rsa-dev
    Port 8080
```

Modifier ~/.gitconfig

```
[url "git@git.webedia-group.net:"]
        insteadOf = https://git.webedia-group.net/
```

Puis
```
go get -insecure git.webedia.group.com/tools/adsgolib
```

