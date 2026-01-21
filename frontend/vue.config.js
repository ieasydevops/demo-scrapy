module.exports = {
  devServer: {
    port: 5000,
    proxy: {
      '/api': {
        target: 'http://localhost:5080',
        changeOrigin: true
      }
    }
  }
}
