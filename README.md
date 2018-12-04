# baseadmin

This "application" is for making the administration and the entire website management more clear and easy with go.

There is one config, one menu, and one roles yml which are containing the settings and related informations and that is not neccesary to place in database.

### Example of config parts ```config.yml```
```
listenport: 8080
```
listenport is saying to the application which port should it listen on obviusly


```
viewdir: view
```
viewdir is the html template files path inside the application directory

```
chiefadmin:
  0:
    email:  istvan.dombi@agl.group
    password: Scl181003
    superdmin: true 
```
chiefadmin is an array of admins will be made by the application if the ```rebuild_structure``` flag is true
```
db:
  environment:
    local:
      host: "host"
      name: "database_name"
      username: "database_user"
      password: "database_password"
```
db is a string map with the ooportunity of setting up multiple environment without removing and modifying the previous one, we will use the key under the ```environment``` to idenfity which environment we want to use
```
  maxidleconns: 20
  maxopenconns: 20
  maxconnlifetimeminutes: 60
```
still inside ```db``` config we have these three option to handle a bit more the mysql pool, ```maxidleconns```, ```maxopenconns```, and ```maxlifetimeminutes```. These are existing options you can research for to understand its working. 
```
environment: local
```
and this is the part we choose our current environment the app should use
```
server:
  readtimeoutseconds: 20
  writetimeoutseconds: 20
  maxrps: 5
  banminutes: 10
  banactive: false
  sessionkey: "baseadmin"
  name: "Base Admin Server"
```
the next parameter is the apps server config, read and write timeout in seconds(```maxrps```), max request per seconds (this is because of defending against hackers, don't know if currently is working or not because was there a proxy problem which occured with this some error, will check about it soon) - it has a related ```banminutes```, obviusly meaning, and a ```banactive``` key which is for activate and deactivate this entire feature.

Also there are ```sessionkey``` which is a string what will be used to encode/decode session content.
The last key here is the ```name``` which is only an informative key what is reachable from response headers if I remember good :D.
```
mode:
  live: false
  debug: true
  rebuild_structure: false
  rebuild_data: false
```
Well, under ```mode``` there are ```live``` which is not sure is used right now, ```debug``` which is for debugging, if you set it true, you will get much more log, ```rebuild_structure``` which is the database rebuild flag. If this is true, the process will remove its tables if these were configured good, and remake them. ```rebuild_data``` is a related process, not for structure but the data.

```
cache:
  enabled: true
  type: "file"
  dir: "view/cache"
```
cache has two types now, ```"file"``` and ```"memory"```. File cache can not store models and values, so if you change from one type to other, maybe it can occur some fail because of this. If you use file cache, there is the ```dir``` option to set the file cache directory. As you see in this example, this is the cache directory under view.
```
adminrouter: admin
```
this is the administration panel access url under our site url. You can find the login to administrative portal on this path. In this example, this is admin, so on localhost it should be accessible via http://localhost:8080/admin url.
```
og:
  url: "OG Url"
  type: website
  title: "Site Default Name"
  description: "Site Default Description"
  image: "/opengraph/default/image.png"
```
these are default opengraph works, can can overwrite from controller, will see soon.
```
ug:
  - value: "usergroup1"
    label: "User Group 1"
    description: "User Group 1 Description"
    default: true
  - value: "usergroup2"
    label: "User Group 2"
    description: "User Group 2 Description"
    default: false
  - value: "admin"
    label: "Admin"
    description: "Admin"
    default: false
```
these are the user groups, ```value```, ```label```, ```description``` and ```default``` option, description not used yet.
```
language:
  allowed:
    - hu
    - en
```
the under language, ```allowed``` key contains the allowed language codes (not iso, just as you see). This is not enough to set language, I will write more about it later.

### Explanation and example of roles ```roles.yml```

```
roles:
  user:
    title: "<strong>All User Role</strong>"
    value: "user/*"
    children:
      list:
        title: "List/Search Users"
        value: "user/list"
      new:
        title: "Add User"
        value: "user/new"
      edit:
        title: "Edit User"
        value: "user/edit"
      delete:
        title: "Delete User"
        value: "user/delete"
  block:
    title: "<strong>All Block Roles</strong>"
    value: "block/*"
    children:
      list:
        title: "List/Search Blocks"
        value: "block/list"
      new:
        title: "Add Block"
        value: "block/new"
      edit:
        title: "Edit Block"
        value: "block/edit"
      delete:
        title: "Delete Block"
        value: "block/delete"
  config:
    title: "<strong>Config</strong>"
    value: "config/*"
    children:
      index:
        title: "Edit config"
        value: "config/index"
```

roles.yml contains every role, it has to be update, because if you want to add a new user, you can chose from the role only are in this file.

under the ```roles``` key, you can see the role groups, fe.: ```user```, ```block```, ```config```
under the groups, there are 3 keys: ```title```, ```value``` and ```children```

the ```title```'s content will be visible on the admin panel when editing user and/or their roles.
the ```value``` contains the group's value, for example user group's value is ```"user/*"``` which means every role are granted to the user who has this role. Does not need to give them children roles under that one by one.

the ```children``` key is the container of all group related subrole. For example, you can make here a new role for allow or deny users to upload image to users, so you make a new role with ```image``` key under the ```user```/```children``` with fe.: ```title: "Add/Edit profile image"``` and ```value: "user/image"```

now, the role is available, and the admin can be set to allow/deny to edit users' images, "later" I will show you how, this is only for explaining how to add a new role to the existing ones.

