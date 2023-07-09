FROM alpine:latest
COPY hyppo .
COPY views /views
COPY assets /assets
COPY public /public
CMD [ "./hyppo"]
