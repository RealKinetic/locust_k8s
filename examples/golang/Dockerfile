# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Start with a base Golang image
FROM golang

MAINTAINER Beau Lyddon <beau.lyddon@realkinetic.com>

# Add the external tasks directory into /example
RUN mkdir example
ADD example_server.go example
WORKDIR example

# Build the example executable
RUN go build example_server.go

# Set server to be executable
RUN chmod 755 example_server

# Expose the required port (8080)
EXPOSE 8080

# Start our example service
ENTRYPOINT ["./example_server"] 
