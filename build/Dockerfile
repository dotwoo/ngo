# FROM ubuntu:18.04 as builder
FROM golang as builder

# RUN apt-get update
# RUN apt-get install -y unzip tar curl sudo tzdata

# Modify timezone to GTM+8
# ENV TZ=Asia/Shanghai
# RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ >/etc/timezone
# RUN dpkg-reconfigure -f noninteractive tzdata

ENV GOPROXY=https://goproxy.cn
ENV CGO_ENABLED=1

# RUN apt-get clean && rm -rf /var/lib/apt/lists/*

WORKDIR /opt/ngo/
COPY . /opt/ngo/

RUN make mod
RUN make

# ## prod
# FROM ubuntu:18.04
FROM prantlf/alpine-glibc
MAINTAINER dotwoo

# RUN apt-get update
# RUN apt-get install -y tzdata

# # Modify timezone to GTM+8
# ENV TZ=Asia/Shanghai
# RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ >/etc/timezone

WORKDIR /opt

COPY --from=builder /opt/ngo/ngo /opt/
COPY --from=builder /opt/ngo/configs/config_sample.yaml /opt/app.yaml

# RUN mkdir -p /data/

EXPOSE 8080
# VOLUME ["/data/"]

CMD ["./ngo", "-c", "./app.yaml"]
