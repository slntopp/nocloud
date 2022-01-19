module.exports = {
  transpileDependencies: [
    'vuetify'
  ],

  publicPath: '/admin',

	chainWebpack: config => {
        config
        .plugin('html')
        .tap(args => {
          args[0].title = 'NoCloud'
          return args
        })
      }
}
