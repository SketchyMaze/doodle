FROM debian:latest
MAINTAINER Noah Petherbridge <root@kirsle.net>
ENV GOPATH /home/builder/go

RUN apt update && apt -y upgrade && \
	apt -y install git zip golang \
	libsdl2-dev libsdl2-ttf-dev make && \
	apt clean

# Create a user to build the packages.
RUN useradd builder -u 1000 -m -G users

# Add the project to the GOPATH
ADD . /home/builder/go/src/git.kirsle.net/SketchyMaze/doodle
RUN chown -R builder:builder /home/builder/go

# Build the app
USER builder
WORKDIR /home/builder/go/src/git.kirsle.net/SketchyMaze/doodle
RUN make setup
CMD ["make", "__docker.dist"]
