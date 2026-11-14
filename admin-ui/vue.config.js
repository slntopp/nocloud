module.exports = {
  transpileDependencies: [
    'vuetify'
  ],

  publicPath: '/admin',

  // source maps add ~25 MB to dist and build time; users never load them
  productionSourceMap: false,

	chainWebpack: config => {
        config
        .plugin('html')
        .tap(args => {
          args[0].title = 'NoCloud'
          return args
        })
      }
}
