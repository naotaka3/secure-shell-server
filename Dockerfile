FROM scratch
COPY secure-shell-server /
ENTRYPOINT ["/secure-shell-server"]
