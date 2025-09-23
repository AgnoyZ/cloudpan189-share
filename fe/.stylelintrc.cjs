module.exports = {
  extends: ['stylelint-config-standard', '@stylistic/stylelint-config'],
  plugins: ['@stylistic/stylelint-plugin'],
  overrides: [
    {
      files: ['**/*.{vue,html}'],
      customSyntax: 'postcss-html',
    },
  ],
  rules: {
    'block-no-empty': true,
    // 允许 Vue SFC 中的 :deep 伪类
    'selector-pseudo-class-no-unknown': [true, { ignorePseudoClasses: ['deep'] }],
    // 允许第三方组件库的类名模式（如 n-card__content）
    'selector-class-pattern': null,
  },
}
