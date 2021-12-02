FROM ubuntu:20.04

# Ubuntu necesseties
RUN apt-get update
RUN apt-get install -y \
    dialog \
    apt-utils

# Set tzdata because it seems to not go automatically
ENV TZ=America/New_York
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

# Coding necesseties
RUN apt-get install -y git
# dunno why but we have to update again
RUN apt-get update

# quality of life
RUN apt-get install -y \
    tmux \
    vim \
    curl \
    neovim \
    unzip

# emp dependencies
RUN apt-get install -y \
    python3-pip \
    software-properties-common \
    build-essential \
    libssl-dev \
    cmake

WORKDIR /root

# copy dotfiles
COPY ./dotfiles/.bashrc .bashrc
COPY ./dotfiles/.vimrc .vimrc
RUN mkdir .config; mkdir .config/nvim
COPY ./dotfiles/init.vim .config/nvim
COPY ./dotfiles/tmux_color_setup.sh tmux_color_setup.sh
COPY ./dotfiles/.tmux.conf .tmux.conf
COPY ./dotfiles/.inputrc .inputrc

# vim-plug
RUN sh -c 'curl -fLo /root/.local/share/nvim/site/autoload/plug.vim --create-dirs \ 
       https://raw.githubusercontent.com/junegunn/vim-plug/master/plug.vim'
RUN nvim +'PlugInstall' +qa

# Install emp-tool
RUN ls
RUN mkdir emp
WORKDIR /root/emp
COPY ./shared/emp_install.py install.py
RUN python3 install.py --install --tool --ot --ag2pc 
# fix things in emp-tool that need fixing
COPY ./shared/changes_to_emp/emp-ag2pc/2pc.h /root/emp/emp-ag2pc/emp-ag2pc/2pc.h
COPY ./shared/changes_to_emp/test/CMakeLists.txt /root/emp/emp-ag2pc/test/CMakeLists.txt
COPY ./shared/changes_to_emp/test/single_execution.h /root/emp/emp-ag2pc/test/single_execution.h
COPY ./shared/changes_to_emp/test/evnotif.cpp /root/emp/emp-ag2pc/test/evnotif.cpp
COPY ./shared/changes_to_emp/test/benchmark.cpp /root/emp/emp-ag2pc/test/benchmark.cpp
WORKDIR /root/emp/emp-ag2pc
RUN rm CMakeCache.txt
RUN make clean
RUN make

WORKDIR /root




