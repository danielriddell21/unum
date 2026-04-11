FROM gcr.io/distroless/static:nonroot
COPY unum /unum
EXPOSE 8080
ENTRYPOINT ["/unum"]
