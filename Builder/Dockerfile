#/*******************************************************************************
# * Copyright 2019 Samsung Electronics All Rights Reserved.
# *
# * Licensed under the Apache License, Version 2.0 (the "License");
# * you may not use this file except in compliance with the License.
# * You may obtain a copy of the License at
# *
# * http://www.apache.org/licenses/LICENSE-2.0
# *
# * Unless required by applicable law or agreed to in writing, software
# * distributed under the License is distributed on an "AS IS" BASIS,
# * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# * See the License for the specific language governing permissions and
# * limitations under the License.
# *
# *******************************************************************************/

FROM ubuntu:18.04

# Prerequisite
RUN apt-get update && apt-get -y upgrade && apt-get install -y --no-install-recommends \
	build-essential \
#	crossbuild-essential-arm64 crossbuild-essential-armhf \
	software-properties-common \
	sudo \
	curl \
	cmake \
	pkg-config \
	linux-libc-dev \
	libc6-dev-i386 \
	libglib2.0-dev \
	gcc-7-multilib g++-7-multilib \
	gcc-aarch64-linux-gnu g++-aarch64-linux-gnu \
	gcc-arm-linux-gnueabi g++-arm-linux-gnueabi \
	lcov \
	libgtest-dev \
	wget \
	zip unzip \
	vim \
	git


# Install C toolchain
RUN	cd /usr/src/gtest && cmake . && make && mv *.a /usr/lib && \
	ln -s /usr/include/asm-generic/ /usr/include/asm


# Install go tools and packages
RUN add-apt-repository ppa:masterminds/glide && apt-get update && apt-get install -y glide
RUN curl -s https://dl.google.com/go/go1.11.5.linux-amd64.tar.gz | tar -v -C /usr/local -xz
# Environment
ENV HOME /home
ENV GOROOT /usr/local/go
RUN mkdir $HOME/go
ENV GOPATH $HOME/go
ENV PATH $PATH:$GOROOT/bin:$GOPATH/bin


# Install Android build env
RUN mkdir -p $HOME/Android
ENV ANDROID_HOME $HOME/Android
ENV ANDROID_NDK_VERSION r19c
ENV ANDROID_NDK_HOME $ANDROID_HOME/android-ndk-${ANDROID_NDK_VERSION}
# Install android-ndk
RUN wget https://dl.google.com/android/repository/android-ndk-${ANDROID_NDK_VERSION}-linux-x86_64.zip && \
	unzip android-ndk-${ANDROID_NDK_VERSION}-linux-x86_64.zip -d $ANDROID_HOME && \
	rm -rf android-ndk-${ANDROID_NDK_VERSION}-linux-x86_64.zip 
ENV PATH $PATH:$ANDROID_NDK_HOME
# Install android-sdk-tools
RUN wget https://dl.google.com/android/repository/sdk-tools-linux-4333796.zip && \
	unzip sdk-tools-linux-4333796.zip -d $ANDROID_HOME && \
	rm -rf sdk-tools-linux-4333796.zip
ENV PATH $PATH:$ANDROID_HOME/tools/bin
RUN apt-get -y install openjdk-8-jdk
RUN yes | sdkmanager "platform-tools" "platforms;android-28"
# Install gradle
RUN wget https://services.gradle.org/distributions/gradle-5.4.1-bin.zip && \
	mkdir /opt/gradle && \
	unzip gradle-5.4.1-bin.zip -d /opt/gradle && \
	rm -rf gradle-5.4.1-bin.zip
ENV PATH $PATH:/opt/gradle/gradle-5.4.1/bin


# Install Tizen build env
RUN echo "deb [trusted=yes] http://download.tizen.org/tools/latest-release/Ubuntu_16.04/ /" >> /etc/apt/sources.list
RUN apt-get update && apt-get install -y \
	gbs mic \
	file \
	libbssolv-perl \
	module-init-tools


# Working directory
WORKDIR $HOME






