# house-web-crawler
A SaaS in GoLang that has a REST API that easily looks for an apartment/house rental contract via a web crawler that collects the latest info from several websites with rental contracts. Read WIKI for License information.

Made by Adam Jacobs and Olle Berglöf.

## Front end repos:
+ https://github.com/osobo/findyourhome-android
+ https://github.com/worldyn/findyourhome-iOS


## Technologies used
+ MongoDB
+ GoQuery
+ PhantomJS
+ Apple Push notification service

## What you need to setup iOS Push:
1. Get an Apple Developer Account.
2. Find your key Id, team Id and add those in apple-setup.go or in a seperate file
3. Add the .p8 file key to the apns.p8 file.
4. Add app bundle in app.js

## What you need to setup HTTPS/SSL
+ Get an SSL-certificate.
+ You should now have a `certificate file` and a `key file`.
+ Make sure they are both in `.pem` format and named `cert.pem` and `key.pem`.
+ Make sure to have the files located in the working directory from where you are running the server.
