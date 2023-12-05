FROM ubuntu:lunar as builder
LABEL dockerfile.version="1"
LABEL software="obitools4"

WORKDIR /
RUN apt update  --fix-missing && apt upgrade -y
RUN apt install -y build-essential 
RUN apt install -y git tcsh bash gawk parallel gettext zlib1g-dev libglib2.0-0
# RUN git clone https://git.metabarcoding.org/org-asm/org-annotate.git
# RUN cd org-annotate/src && make && cd ../..
# RUN cd /org-annotate/data/its/ITSx_db/HMMs && \
#     rm *.h3* && \
#     for f in *.hmm ; do /org-annotate/ports/i386-linux/bin/hmmpress $f ; done

RUN apt install -y python3-pip python3-dev python3-venv
RUN git clone https://git.metabarcoding.org/org-asm/org-asm.git
RUN cd org-asm 
RUN python3 -m venv ../org-assembler 
RUN bash -c '../org-assembler/bin/pip3 install -r /org-asm/requirements.txt'
RUN bash -c '../org-assembler/bin/python3 setup.py install --no-serenity'
RUN cd ..


# FROM ubuntu:lunar as phyloskims
# WORKDIR /
# RUN apt update  --fix-missing && apt upgrade -y
# RUN apt install -y tcsh bash gawk parallel zlib1g libglib2.0-0
# COPY --from=builder /org-annotate /org-annotate
# RUN rm -rf /org-annotate/src
# COPY --from=builder /org-assembler /org-assembler
# RUN mkdir -p /data
