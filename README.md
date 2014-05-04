Gobble
======

This is a simple blogging engine written in Go.  Its features are:

 - Works on any platform that Go can build for.
 - Does not require a database.
 - Syntax highlighting via [Rainbow][1].
 - Comment spam detection via [Akismet][2].
 - Comment spam prevention via [reCAPTCHA][3].
 - Easy to install.
 - Fast.
 - Python 3 script to convert from an XML WordPress export to Gobble format.
 - Configurable via JSON file.
 - Posts are stored on the file system.
 - Posts and comments are written in Markdown.
 - Search.
 - Archives list.
 - Paging.
 - Tagging.
 - RSS feed.
 - Simple re-theming.

  [1]: http://craig.is/making/rainbows
  [2]: http://akismet.com
  [3]: http://www.google.com/recaptcha


Writing Posts
-------------

All posts are stored in the gobble/posts directory.  All posts are written in
Markdown, with metadata included at the top of the post giving the publish date,
tags, etc.  The format is identical to that used by [Scriptogram][4].

  [4]: http://scriptogr.am

To write a post, create a new file in the posts directory.  Call it whatever you
like, but ensure it has the extension ".md".  Here's an example:


    Title: My First Gobble Post
    Date: 2013-02-02 01:18:36
    Tags: helloworld, fristpost

    This is my first Gobble post!


Save the file and start Gobble.  Your post should now appear.  Clicking on the
"Tags" link in the navigation menu will show the two tags, and the "Archives"
page will show this new post's title and publish date.


Media Files
-----------

Media files, such as images, sounds and others, can be served from the folder
defined in the `mediaPath` config setting.  Media links take this format:

    /media/path/to/your/file

Files can be organised however you want.  I personally organise my files like
WordPress and use a directory structure like this:

    /media/year/month/day/file

For example:

    /media/2014/01/26/Image.png

To link to that image from a post, this is the Markdown syntax:

    ![Image name](/media/2014/01/26/Image.png)


Other Static Files
------------------

Other static files which support the blog itself, rather than content, should be
stored in the path defined by the `staticFilePath` config setting.  These files
include the robots.txt file, favicon.ico file, etc and are served from a user-
defined URL.  The files have to be added to the config file in order for Gobble
to serve them.

Here's a partial example of a config file that will serve `robots.txt` and
`favicon.ico` from the `./files` directory:

    {
        ...
        "staticFilePath": "./files",
        "staticFiles": {
            "/favicon.ico": "favicon.ico",
            "/robots.txt": "robots.txt"
        }
    }

In the "staticFiles" dictionary, the key represents the URL of the file and the
value represents the path to the file relative to the `staticFilePath` value.
In this example, the URL `/favicon.ico` serves the file located at
`./files/favicon.ico`.


Tagging
-------

Tags are specified as a list of words in a post's metadata block, separated by
commas.  They should be lower-case, but Gobble will automatically convert them
to lower-case for you should you enter some with upper-case characters.

Hash characters in tags are stripped out.  Tags are included in URLs, so hash
characters would prevent the browser from accessing the correct page.


Comments
--------

Comments are stored in a folder in the posts directory that has the same name
as the post's Markdown file.  For example, a post called "my-first-post.md" will
store its comments in a folder called "my-first-post/comments".

Comments can be disabled on a post-by-post basis by using the "DisallowComments"
metadata tag:

    DisallowComments: true

Omitting the tag or using any value other than "true" will enable comments.
This functionality is provided mainly as a way to stop spam bots that latch on
to a particular post and repeatedly manage to bypass the other spam protections.

Comments can also be disabled by setting a value for the "commentsOpenForDays"
configuration property.  Setting this to a value other than `0` will cause
comments to be disabled after the specified number of days.  This is useful both
for blocking spam and for preventing discussion of ancient posts.

Other spam protection is implemented via Akismet and reCAPTCHA.  Both services
are enabled automatically if their keys are provided in the config file.


Theming
-------

Gobble supports multiple themes, which can be found in the gobble/themes
directory.  Each theme consists of image, css and templates folders.  All are
designed to be standards compliant and easy to edit.  Each theme can also
include a "favicon.ico" file and a "robots.txt" file.  Both files should be
placed in the root folder of the theme:

    - themes/
      - grump/
        - css/
        - img/
        - templates/
        - favicon.ico
        - robots.txt

When editing templates, take care not to disturb content between {{brackets}}.
These are part of Go's templating system and function as placeholders for
content generated by Gobble.  Changing these may break the templates.


Post Caching
------------

Rather than use a database, which would increase the complexity of Gobble and
its installation, Gobble stores all of its posts on the file system.  When
Gobble is running it caches the entire content of the blog in memory.  This may
sound a little excessive, but simianzombie.com consists of 280,000 words spread
over 480 posts and 912 comments, and uses just 2MB of disk space.  Storing the
posts in RAM makes retrieving and searching them extremely fast.

The cache is reloaded whenever the content of the posts directory changes.


Installation
------------

The easiest way to install Gobble is via the command line.  Assuming you have Go
installed and configured correctly:

    go get github.com/ant512/gobble
    go get github.com/bmizerany/pat
    go get github.com/dpapathanasiou/go-recaptcha
    go get github.com/howeyc/fsnotify
    go get github.com/russross/blackfriday
    cd $GOPATH/src/github.com/ant512/gobble
    go build
    ./gobble

Gobble wil now be available at http://localhost:8080.


Ubuntu Gobble Service
---------------------

If you are deploying to an Ubuntu server, you can set up Gobble as a service to
run at system startup:

 - cd /etc/init
 - sudo nano gobble.conf
 - Insert the following text:

    description     "gobble web server"

    start on startup

    chdir path_to_gobble/gobble
    exec ./gobble

 - sudo service gobble start

To stop Gobble:

 - sudo service gobble stop


Startup Options
---------------

The only command line argument currently available is `-config`, which allows
you to specify a different config file for Gobble to load:

    ./gobble -config ./gobble.conf


Configuration
-------------

Gobble's configuration is changed by editing the gobble.conf file.  You can
specify a different file to use via the command line, and therefore have
multiple Gobble servers running simultaneously.

The default config file looks like this:

    {                                                 
        "name": "Gobble",
        "description": "Blogging Engine",
        "address": "http://simianzombie.com",
        "port": 8080,
        "postPath": "./posts",
        "mediaPath": "./media",
        "themePath": "./themes",
        "theme": "grump",
        "commentsOpenForDays": 0,
        "akismetAPIKey": "",
        "recaptchaPublicKey": "",
        "recaptchaPrivateKey": "",
        "staticFilePath": "./files",
        "staticFiles": { }
    }

The config file is a JSON document.  When editing the file, ensure that you
respect the JSON conventions.

The options are defined as follows:

 - name:                the site's name.
 - description:         the site's description, which appears on the RSS feed.
 - address:             the site's address, which appears on the RSS feed and is
                        sent to Akismet for comment validation.
 - port:                the port on which Gobble should listen.
 - postPath:            the path to the posts directory.
 - mediaPath:           the path to the media directory.
 - themePath:           the path to the themes directory.
 - theme:               the theme to use.
 - commentsOpenForDays: the number of days that comments can be added to a post
                        after its publish date (0 means "forever").
 - akismetAPIKey:       the key used to check comments for spam (leave it blank
                        if you don't want to use Akismet).
 - recaptchaPublicKey:  the key used to ensure the commenter isn't a bot (leave
                        it blank if you don't want to use reCAPTCHA).
 - recaptchaPrivateKey: the key used to ensure the commenter isn't a bot (leave
                        it blank if you don't want to use reCAPTCHA).
 - staticFilePath:      the path to the files directory, which contains the
                        robots.txt, favicon.ico, and others.
 - staticFiles:         a dictionary of files to serve from the files directory;
                        the key is the URL, and the value is the filename.

Note that missing configuration values will be given the defaults.


Nginx
-----

Nginx can be used as a proxy to redirect traffic to the Gobble server.  Here's
an example server block:

    server {
            listen 80;
            server_name example.com;
            access_log /var/log/nginx/example.com.access.log;
            location / {
                proxy_pass http://127.0.0.1:8080;
            }
    }


Libraries
---------

Gobble uses a handful of libraries to do its thing:

 - [http://craig.is/making/rainbows][5]
 - [http://github.com/bmizerany/pat][6]
 - [http://github.com/dpapathanasiou/go-recaptcha][7]
 - [http://github.com/howeyc/fsnotify][8]
 - [http://github.com/russross/blackfriday][9]

  [5]: http://craig.is/making/rainbows
  [6]: http://github.com/bmizerany/pat
  [7]: http://github.com/dpapathanasiou/go-recaptcha
  [8]: http://github.com/howeyc/fsnotify
  [9]: http://github.com/russross/blackfriday
