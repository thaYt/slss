import postcss from 'rollup-plugin-postcss';

export default {
	input: 'src/main.js',
	plugins: [postcss()]
};