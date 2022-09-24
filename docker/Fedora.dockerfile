FROM fedora:latest
MAINTAINER Noah Petherbridge <root@kirsle.net>
ENV GOPATH /home/builder/go

# Update all the software and then install Go, git, SDL2 and other dependencies
RUN dnf -y update && \
	dnf -y install git zip golang SDL2-devel SDL2_ttf-devel make && \
	dnf clean all

# Create a user to build the packages.
RUN useradd builder -u 1000 -m -G users

# Add the project to the GOPATH
ADD . /home/builder/go/src/git.kirsle.net/SketchyMaze/doodle
WORKDIR /home/builder/go/src/git.kirsle.net/SketchyMaze/doodle
RUN chown -R builder:builder /home/builder/go

# Build the app as the `builder` user
USER builder
RUN make setup
CMD ["make", "__docker.dist"]
