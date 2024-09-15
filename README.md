A small and simple program for having your Pleroma notifications printed live into your terminal. Mainly meant for use in acme [plan 9 text editor]

## Building
Just type make, there aren't any dependencies other than go.

## Using
Create a file with your instance and your account api token [**not your password**] seperated by a newline. It should look like this:
```
freecumextremist.com
chodeington
```

Then simply invoke the notifs binary with the name of your new file as an argument: `./notifs details.txt`

You can find your token in any request to the mastodon api from devtools [look in the headers, value should start with "Bearer" don't paste the Bearer part though], you can also make a new token just for this program by manually probing the mastodon api tediously, or by just using one of those websites that just let you fill out a form and do the process automatically. 
