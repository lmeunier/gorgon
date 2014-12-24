Gorgon
======

Overview
--------

Gorgon is a `Persona/BrowserId <https://persona.org/>`_ Identity Provider (IdP)
written with the `Go Programming Language <http://golang.org/>`_.

Gorgon is yet in active development state. Do **NOT** use it in production.

Install
-------

You can install Gorgon using a released version (download a tarball from the
`Releases page on Github <https://github.com/lmeunier/gorgon/releases>`_), or
you can `build Gorgon from sources <#build>`_.

- extract the tarball

.. code:: bash

   tar xaf gorgon-0.1.0.tar.gz
   cd gorgon-0.1.0

- create a private and a public keys

.. code:: bash

   openssl genrsa -out private-key.pem 2048
   openssl rsa -in private-key.pem -pubout > public-key.pem

- copy and edit the default configuration file

.. code:: bash

   cp gorgon.ini.example gorgon.ini
   $EDITOR gorgon.ini

Configure
---------

Gorgon needs a configuration file. By default, it will try to read the file
``gorgin.ini`` in the current folder, you can change the path to the
configuration file with the ``-c`` flag. For example:

.. code:: bash

   ./gorgon -c /etc/gorgon/my_gorgon_config.ini

The configuration file is a classic INI file parsed with `go-ini
<https://github.com/vaughan0/go-ini#file-format>`_.

Test Authenticator
~~~~~~~~~~~~~~~~~~

The Test Authenticator let you define a global password, this password will be
used for all authentication attempts (you can authenticate as any user with
this password). Do **NOT** use this authentication method in production.

.. code:: ini

   [global]
   ...
   auth = test

   [auth:test]
   global_password = myverysecretpassword

IMAP Authenticator
~~~~~~~~~~~~~~~~~~

The IMAP Authenticator uses an IMAP server to authenticate users. If the IMAP
server advertise the ``STARTTLS`` capability, the connection switches to TLS.
The username (email address) and password are sent without modification to the
IMAP server.

.. code:: ini

   [global]
   ...
   auth = imap

   [auth:imap]
   server = imap.example.com

Run
---

Once Gorgon is `installed <#install>`_ and `configured <#configure>`_, you are
ready to run it.  To start Gorgon, you just have to invoke the ``./gorgon``
command in the folder where Gorgon is installed.

Gorgon will not daemonize itself. To run Gorgon as a background process, you
must use a tool like `Supervisor <http://supervisord.org/>`_ or `systemd
<http://freedesktop.org/wiki/Software/systemd/>`_.

Once started, Gorgon will listen for HTTP requests on the ``interface:port``
defined in the configuration file. It's up to you to configure your webserver
to redirect HTTP requests to Gorgon.

Serve
-----

Every Persona IdP must be served:

- over HTTPS
- from the exact host part of the email address, not a subdomain

For example, if your email address is ``alice@example.com``, you must configure
your webserver to redirect every requests for
``https://example.com/.well-known/browserid`` (and everything under this URL)
to Gorgon.

Here are example configurations for common webservers.

Nginx
~~~~~

.. code::

  server {
    listen [::]:443;
    server_name "example.com";
    ssl on;
    ssl_certificate /path/to/example.com.crt;
    ssl_certificate_key /path/to/private.key;

    location /.well-known/browserid {
      # Gorgon is listening on port 5000
      proxy_pass http://127.0.0.1:5000;
    }
  }

Apache
~~~~~~

.. code::

  <VirtualHost *:443>
    ServerName example.com
    SSLEngine On
    SSLCertificateFile /path/to/example.com.crt
    SSLCertificateKeyFile /path/to/private.key

    <Location /.well-known/browserid>
      # Gorgon is listening on port 5000
      ProxyPass / http://127.0.0.1:5000/
      ProxyPassReverse / http://127.0.0.1:5000/
    </Location>
  </VirtualHost>


Build
-----

Gorgon uses `Gox <https://github.com/mitchellh/gox>`_ to build and cross
compile the application for multiple platforms. Before trying to build Gorgon,
make sure you have a working Gox installation.

By default, the ``Makefile`` will build Gorgon for common platforms
(linux/darwin/*bsd). You can modify the ``OSARCHS`` variable in the
``Makefile`` to add or remove platforms.

- initialize a workspace directory and set ``GOPATH`` and ``PATH`` accordingly

.. code:: bash

    mkdir -p "$HOME/gorgon/gopath"
    export GOPATH="$HOME/gorgon/gopath"
    export PATH="$GOPATH/bin:$PATH"

- install Gorgon sources

.. code:: bash

    go get -d github.com/lmeunier/gorgon

- build Gorgon

.. code:: bash

    cd "$GOPATH/src/github.com/lmeunier/gorgon"
    make install_deps
    make build

The ``build`` target of the Makefile will create a ``gorgon`` executable file
for each platform listed in the ``OSARCHS`` variable in the ``Makefile``,
these files are created in the ``build/`` folder.

- create tarballs

.. code:: bash

   make dist

The ``dist`` target of the Makefile will create a tarball archive for each
platform listed in the ``OSARCHS`` variable in the ``Makefile`` in the
``dist/`` folder. You can use these tarballs to `install Gorgon <#install>`_.

Credits
-------

Gorgon is maintained by `Laurent Meunier <http://www.deltalima.net/>`_.

Licence
-------

Gorgon is Copyright (c) 2014 Laurent Meunier. It is free software, and may be
redistributed under the terms specified in the LICENSE file (a 3-clause BSD
License).
