FROM ubuntu:latest
MAINTAINER Noah Petherbridge <root@kirsle.net>
ENV GOPATH /home/builder/go

# Update all the software and then install Go, git, SDL2 and other dependencies
RUN apt update && \
	apt -y upgrade && \
	apt -y install git zip golang libsdl2-dev libsdl2-ttf-dev make && \
	apt clean

# Create a user to build the packages.
RUN useradd builder -u 1000 -m -G users

# Add the project to the GOPATH
ADD . /home/builder/go/src/git.kirsle.net/apps/doodle
WORKDIR /home/builder/go/src/git.kirsle.net/apps/doodle
RUN chown -R builder:builder /home/builder/go

# Build the app as the `builder` user
USER builder
RUN make setup
CMD ["make", "__docker.dist"]
