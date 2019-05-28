# Introduction Baseadmin

This "application" is for making the administration and the entire website management more clear and easy with go.
This works with mysql database and driver (https://github.com/go-sql-driver/mysql), gorp (https://github.com/go-gorp/gorp) package, and fasthttp (https://github.com/valyala/fasthttp) routing and request handling.

## Getting started

First thing you need is to clone the app into the project directory. 

As long as the repo is not works as a callable outer package, you have to work with, and extend the cloned codebase.

One way is enter to projects root directory (go source directoy ```go/src``` directory) from terminal, and type ```git clone https://github.com/dombiistvan28/baseadmin.git mynewdirectory```. This will clone the project into your new ```mynewdirectory``` directory. You can clone only into an empty directory. Now you still have to replace all "baseadmin" string to "mynewdirectory" because the package and directory name must be the same, or the GO will not recognize it.

Other way is to download archive, and extract into your project's directory.

There is one config, one menu, and one roles yml which are containing the settings and related informations and that is not neccesary to place in database.

### Example of config parts ```config.json```
```
"listenPort": "8080",
```
listenPort is saying to the application which port should it listen on obviusly


```
"viewDirectory": "view",
```
viewdir is the html template files path inside the application directory

```
"chiefAdmin":[
  {
    "email":  "dombiistvan28@gmail.com",
    "password": "BaseAdmin20060102",
    "superadmin": true
  }
]
```
chiefAdmin is an array of admins will be made by the application if the ```rebuildStructure``` flag is ```true```

```
"db":{
  "environment":{
    "local":{
      "host": "host",
      "name": "database_name",
      "username": "database_user",
      "password": "database_password"
    }
  }
}
```
db is an object with the oportunity of setting up multiple environment without removing and modifying the previous one. we will use later the key under the ```environment``` to idenfity which environment we want to use

```
  "maxIdleCons": 20,
  "maxOpenCons": 20,
  "maxConLifetimeMinutes": 60,
```
still inside ```db``` config we have these three option to configure mysql pool, ```maxIdleCons```, ```maxOpenCons```, and ```maxLifetimeMinutes```. These are existing configurations, you can search for to understand how it is working.

```
"environment": "local",
```
and this is the part we choose our current (currently only database) environment the app should use

```
"server":{
  "readTimeoutSeconds": 20,
  "writeTimeoutSeconds": 20,
  "maxRps": 5,
  "banMinutes": 10,
  "banActive": false,
  "sessionKey": "baseadmin",
  "name": "Base Admin Server"
}
```
the next parameter is the app's server config, read (```readTimeoutSeconds```) and write (```writeTimeoutSeconds```) timeout in seconds, max request per seconds (```maxRps```: this is because of defending against hackers, don't know if currently is working or not because there was a proxy problem which occured with this some error, will check about it soon) - it has a related ```banMinutes``` (obviusly meaning), and a ```banActive``` key which is for activate and deactivate this entire feature.

Also there are ```sessionKey``` which is a string what will be used to encode/decode session content.
The last key here is the ```name``` which is only an informative key what is reachable from response headers if I remember good :D.

```
"mode":{
  "live": false,
  "debug": true,
  "rebuildStructure": false,
  "rebuildData": false,
}
```
Well, under ```mode``` there are 
```live``` which is not sure is used right now
```debug``` which is for debugging, if you set it true, you will get much more log
```rebuildStructure``` which is the database rebuild flag: If this is true, the process will remove its tables if these were configured good, and remake them
```rebuildData``` is similar to rebuild_strucure process, not for structure but the data

```
"cache":{
  "enabled": true,
  "type": "file",
  "directory": "view/cache"
}
```
cache has two types now, ```"file"``` and ```"memory"```. File cache can not store models and values, so if you change from one type to other, maybe it can occur some fail because of this. If you use file cache, there is the ```directory``` option to set the file cache directory. As you see in this example, this is the cache directory under view.

```
"adminRouter": "admin",
```
this is the administration panel access url under our site url. You can find the login to administrative portal on this path. In this example, this is admin, so on localhost it should be accessible via http://localhost:8080/admin url.

```
"openGraph":{
  "url": "OG Url",
  "type": "website",
  "title": "Site Default Name",
  "description": "Site Default Description",
  "image": "/opengraph/default/image.png",
}
```
these are default opengraph works, can can overwrite from controller, will see soon.

```
"userGroup":[
  {
    "value": "usergroup1",
    "label": "User Group 1",
    "description": "User Group 1 Description",
    "default": true
  },{
    "value": "usergroup2",
    "label": "User Group 2",
    "description": "User Group 2 Description",
    "default": false
  },{
    "value": "admin",
    "label": "Admin",
    "description": "Admin",
    "default": false
  }
]
```
```value``` is the saved value
```label``` is the select option string
```description``` is not used at the moment
```default``` is for spefify the default to save to the new users

```
"language":{
  "allowed":[
    "hu",
    "en"
  ]
}
```
under language, ```allowed``` key contains the allowed language codes (notISO, just as you see). This is not enough to set language, I will write more about it later.

### Explanation and example of roles ```roles.json```

```
"roles": {
    "user": {
      "title": "User Roles (all)",
      "value": "user/*",
      "children": {
        "list": {
          "title": "List/Search Users",
          "value": "user/list"
        },
        "edit": {
          "title": "Add/Edit User",
          "value": "user/edit"
        },
        "delete": {
          "title": "Delete User",
          "value": "user/delete"
        }
      }
    },
  "block": {
      "title": "Block Roles (all)",
      "value": "block/*",
      "children": {
        "list": {
          "title": "List/Search Blocks",
          "value": "block/list"
        },
        "edit": {
          "title": "Add/Edit Block",
          "value": "block/edit"
        },
        "delete": {
          "title": "Delete Block",
          "value": "block/delete"
        }
      }
    },
  "config": {
      "title": "Config Roles (all)",
      "value": "config/*",
      "children": {
        "index": {
          "title": "Edit config",
          "value": "config/index"
        }
      }
    }
```

roles.json contains every role, it has to be update, because if you want to add a new user, you can chose from the role only are in this file. The roles structure is a json object so ever group (==key) must be unique.

under the ```roles``` key, you can see the role groups, fe.: ```user```, ```block```, ```config```
under the groups, there are 3 keys: ```title```, ```value``` and ```children```

the ```title```'s content will be visible on the admin panel when editing user and/or their roles.
the ```value``` contains the group's value, for example user group's value is ```"user/*"``` which means every role are granted to the user who has this role. Does not need to give them children roles under that one by one.

the ```children``` key is the container of all group related subrole. For example, you can make here a new role for allow or deny users to upload image to users, so you make a new role with ```image``` key under the ```user```/```children``` with fe.: ```title: "Add/Edit profile image"``` and ```value: "user/image"```

now, the role is available, and the admin can be set to allow/deny to edit users' images, "later" I will show you how, this is only for explaining how to add a new role to the existing ones.

### Explanation and example of menu tree ```menu.json```

```
"menu": [
    {
      "label": "User",
      "group": "user",
      "url": "user/index",
      "icon": "fa fa-user",
      "visibility": "*",
      "children": [
        {
          "label": "Log In",
          "url": "user/login",
          "visibility": "!@",
          "icon": "fa fa-user"
        },
        {
          "label": "List",
          "url": "user/index",
          "visibility": "user/list",
          "icon": "fa fa-list"
        },
        {
          "label": "Add New",
          "url": "user/new",
          "visibility": "user/new",
          "icon": "fa fa-plus"
        }
      ]
    },
    ...
```
As you see the ```menu.json``` is an array of json objects.
Every menu group item, has ```label```,```group```,```url```,```icon```, ```visibility``` and ```children``` keys.

```label``` is obvious, it is readable in the menu tree 
```group``` is for the role group. The purpose of it is when you dont have any of the group's subrole, we can hide it at all
```url``` is for the route, it will refer to
```visibility``` is the role we define the user has to have to see this menupoint
```icon``` is just a display bootstrap icon in the menu tree

This ```json``` still not authenticate, just hide/show urls and menupoints. Without specify the accessibility in the controllers (soon) the user can reach the action from url.

There are static roles also you can define in menu, and also to the actions later.
The roles are the following:
```-```: noone
```*```: anyone
```!@```: not logged in user (can be admin or simple user also)
```@```: logged in user (can be admin or simple user also)
```@a```: logged in admin
```@!a```: logged in not admin
```@sa```: logged in superadmin

Easy to add more, and plan to do in the future :) 
