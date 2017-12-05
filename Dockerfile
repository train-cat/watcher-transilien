FROM eraac/golang

ADD watcher-transilien /watcher-transilien

CMD ["/watcher-transilien", "-config", "/config/config.json"]

