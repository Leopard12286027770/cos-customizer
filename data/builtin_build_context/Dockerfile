FROM ubuntu:18.04 
WORKDIR /
RUN apt-get update \
	&& apt-get install -y --no-install-recommends cryptsetup-bin \
	&& rm -rf /var/lib/apt/lists/*
CMD ["veritysetup"]
