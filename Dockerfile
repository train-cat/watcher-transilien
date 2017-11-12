FROM eraac/golang

ADD sniffer-transilien /sniffer-transilien

CMD ["/sniffer-transilien", "-config", "/config/config.json"]

