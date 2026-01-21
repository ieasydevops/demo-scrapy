module.exports = {
  devServer: {
    port: 5001,
    proxy: {
      '/api': {
        target: 'http://localhost:5080',
        changeOrigin: true
      }
    }
  }
}
