<template>
	<div class="service-create-view pa-4">
		<div class="page__title">
			Create service
		</div>
		<service-create/>
	</div>
</template>

<script>
import serviceCreate from '@/components/service-create'

export default {
	name: 'serviceCreate-view',
	components: {
		'service-create': serviceCreate
	}
}
</script>

<style scoped lang="scss">
.page__title{
	color: var(--v-primary-base);
	font-weight: 400;
	font-size: 32px;
	font-family: "Quicksand", sans-serif;
	line-height: 1em;
	margin-bottom: 10px;
}
</style>