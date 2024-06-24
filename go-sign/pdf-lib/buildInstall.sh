cmake --build build -j $(nproc)
sudo cmake --build build --target install -- -j $(nproc)