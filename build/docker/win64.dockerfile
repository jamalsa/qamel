FROM radhifadlillah/qamel:linux as linux

# ========== END OF LINUX ========== #

FROM ubuntu:16.04 as base

RUN apt-get -qq update && \
    apt-get -qq -y install software-properties-common apt-transport-https

RUN apt-key adv --keyserver keyserver.ubuntu.com --recv-keys 86B72ED9
RUN add-apt-repository 'deb [arch=amd64] https://pkg.mxe.cc/repos/apt bionic main'
RUN apt-get -qq update && \
    apt-get -qq -y install mxe-x86-64-w64-mingw32.shared-qt5

# ========== END OF BASE ========== #

FROM ubuntu:16.04

ENV HOME /home/user
ENV GOPATH $HOME/go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH

# Install ca-certificates which might be needed by Go proxy
RUN apt-get -qq update && \
    apt-get -qq -y install ca-certificates git

# Copy Go and Qamel from linux
COPY --from=linux /usr/local/go /usr/local/go
COPY --from=linux $GOPATH/bin $GOPATH/bin
COPY --from=linux $GOPATH/src/github.com/jamalsa/qamel $GOPATH/src/github.com/jamalsa/qamel

# Copy MXE from base
COPY --from=base /usr/lib/mxe /usr/lib/mxe

# Create profile for Qamel
RUN mkdir -p $HOME/.config/qamel
RUN printf '%s %s %s %s %s %s %s %s %s %s %s %s\n' \
    '{"default":{' \
    '"OS":"windows",' \
    '"Arch":"amd64",' \
    '"Static":false,' \
    '"Qmake":"/usr/lib/mxe/usr/x86_64-w64-mingw32.shared/qt5/bin/qmake",' \
    '"Moc":"/usr/lib/mxe/usr/x86_64-w64-mingw32.shared/qt5/bin/moc",' \
    '"Rcc":"/usr/lib/mxe/usr/x86_64-w64-mingw32.shared/qt5/bin/rcc",' \
    '"Gcc":"/usr/lib/mxe/usr/bin/x86_64-w64-mingw32.shared-gcc",' \
    '"Gxx":"/usr/lib/mxe/usr/bin/x86_64-w64-mingw32.shared-g++",' \
    '"Windres":"/usr/lib/mxe/usr/bin/x86_64-w64-mingw32.shared-windres",' \
    '"Objdump":"/usr/lib/mxe/usr/bin/x86_64-w64-mingw32.shared-objdump"' \
    '}}' > $HOME/.config/qamel/config.json

# Build app
ENTRYPOINT ["qamel", "build"]