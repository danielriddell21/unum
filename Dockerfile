FROM gcr.io/distroless/static-debian12:nonroot
COPY unum /unum
EXPOSE 8080
ENTRYPOINT ["/unum"]
