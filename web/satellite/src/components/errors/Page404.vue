// Copyright (C) 2019 Storj Labs, Inc.
// See LICENSE for copying information.

<template src="./page40X.html" />

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator';

import LogoIcon from '@/../static/images/logo.svg';
import MainIcon from '@/../static/images/errors/404.svg';

// @vue/component
@Component({
    components: {
        LogoIcon,
        MainIcon,
    },
})
// 404 component on page not found error
export default class Page404 extends Vue {}
</script>

<style scoped lang="scss">
    .error-container {
        position: fixed;
        top: 0;
        left: 0;
        z-index: 1000;
        width: 100%;
        height: 100vh;
        background-color: #fff;
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
        -webkit-touch-callout: none;
        user-select: none;

        &__title {
            font-family: 'font_regular', sans-serif;
            margin-bottom: 60px;
            font-size: 32px;
        }
    }

    .logo {
        display: flex;
        align-items: center;
        position: absolute;
        top: 87px;
        left: 100px;
    }

    .text {
        margin-left: 15px;
    }
</style>
