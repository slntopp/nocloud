module.exports = {
  module: {
    rules: [
      {
        test: /\.sass$/,
        use: [
          'vue-style-loader',
          'css-loader',
          {
            loader: 'sass-loader',
            options: {
              additionalData: "@import '@/styles/globalStyles.scss'"
            },
          },
        ],
      },
      {
        test: /\.scss$/,
        use: [
          'vue-style-loader',
          'css-loader',
          {
            loader: 'sass-loader',
            options: {
              additionalData: "@import '@/styles/globalStyles.scss';"
            },
          },
        ],
      },
    ],
  },
}