FROM alpine:latest
COPY hyppo .
COPY views /views
COPY assets /assets
CMD [ "./hyppo"]
