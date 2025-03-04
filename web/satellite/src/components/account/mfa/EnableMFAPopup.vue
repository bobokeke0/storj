// Copyright (C) 2021 Storj Labs, Inc.
// See LICENSE for copying information.

<template>
    <div class="enable-mfa">
        <div class="enable-mfa__container">
            <h1 class="enable-mfa__container__title">Two-Factor Authentication</h1>
            <p v-if="isScan" class="enable-mfa__container__subtitle">
                Scan this QR code in your favorite TOTP app to get started.
            </p>
            <p v-if="isEnable" class="enable-mfa__container__subtitle max-width">
                Enter the authentication code generated in your TOTP app to confirm your account is connected.
            </p>
            <p v-if="isCodes" class="enable-mfa__container__subtitle">
                Save recovery codes.
            </p>
            <div v-if="isScan" class="enable-mfa__container__scan">
                <h2 class="enable-mfa__container__scan__title">Scan this QR Code</h2>
                <p class="enable-mfa__container__scan__subtitle">Scan the following QR code in your OTP app.</p>
                <div class="enable-mfa__container__scan__qr">
                    <canvas ref="canvas" class="enable-mfa__container__scan__qr__canvas" />
                </div>
                <p class="enable-mfa__container__scan__subtitle">Unable to scan? Use the following code instead:</p>
                <p class="enable-mfa__container__scan__secret">{{ userMFASecret }}</p>
            </div>
            <div v-if="isEnable" class="enable-mfa__container__confirm">
                <h2 class="enable-mfa__container__confirm__title">Confirm Authentication Code</h2>
                <ConfirmMFAInput :on-input="onConfirmInput" :is-error="isError" />
            </div>
            <div v-if="isCodes" class="enable-mfa__container__codes">
                <h2 class="enable-mfa__container__codes__title max-width">
                    Please save these codes somewhere to be able to recover access to your account.
                </h2>
                <p
                    v-for="(code, index) in userMFARecoveryCodes"
                    :key="index"
                >
                    {{ code }}
                </p>
            </div>
            <div class="enable-mfa__container__buttons">
                <VButton
                    class="cancel-button"
                    label="Cancel"
                    width="50%"
                    height="44px"
                    is-white="true"
                    :on-press="toggleModal"
                />
                <VButton
                    v-if="isScan"
                    label="Continue"
                    width="50%"
                    height="44px"
                    :on-press="showEnable"
                />
                <VButton
                    v-if="isEnable"
                    label="Enable"
                    width="50%"
                    height="44px"
                    :on-press="enable"
                    :is-disabled="!confirmPasscode || isLoading"
                />
                <VButton
                    v-if="isCodes"
                    label="Done"
                    width="50%"
                    height="44px"
                    :on-press="toggleModal"
                />
            </div>
            <div class="enable-mfa__container__close-container" @click="toggleModal">
                <CloseCrossIcon />
            </div>
        </div>
    </div>
</template>

<script lang="ts">
import QRCode from 'qrcode';
import { Component, Prop, Vue } from 'vue-property-decorator';

import ConfirmMFAInput from '@/components/account/mfa/ConfirmMFAInput.vue';
import VButton from '@/components/common/VButton.vue';

import CloseCrossIcon from '@/../static/images/common/closeCross.svg';

import { USER_ACTIONS } from '@/store/modules/users';

import { AnalyticsHttpApi } from '@/api/analytics';
import { AnalyticsEvent } from '@/utils/constants/analyticsEventNames';

// @vue/component
@Component({
    components: {
        ConfirmMFAInput,
        CloseCrossIcon,
        VButton,
    },
})
export default class EnableMFAPopup extends Vue {
    @Prop({default: () => () => false})
    public readonly toggleModal: () => void;

    public readonly qrLink =
        `otpauth://totp/${encodeURIComponent(this.$store.getters.user.email)}?secret=${this.userMFASecret}&issuer=${encodeURIComponent(`STORJ ${this.satellite}`)}&algorithm=SHA1&digits=6&period=30`;
    public isScan = true;
    public isEnable = false;
    public isCodes = false;
    public isError = false;
    public isLoading = false;
    public confirmPasscode = '';

    public $refs!: {
        canvas: HTMLCanvasElement;
    };

    private readonly analytics: AnalyticsHttpApi = new AnalyticsHttpApi();

    /**
     * Mounted lifecycle hook after initial render.
     * Renders QR code.
     */
    public async mounted(): Promise<void> {
        await QRCode.toCanvas(this.$refs.canvas, this.qrLink);
    }

    /**
     * Toggles view to Enable MFA state.
     */
    public showEnable(): void {
        this.isScan = false;
        this.isEnable = true;
    }

    /**
     * Toggles view to MFA Recovery Codes state.
     */
    public async showCodes(): Promise<void> {
        await this.$store.dispatch(USER_ACTIONS.GENERATE_USER_MFA_RECOVERY_CODES);

        this.isEnable = false;
        this.isCodes = true;
    }

    /**
     * Sets confirmation passcode value from input.
     */
    public onConfirmInput(value: string): void {
        this.isError = false;
        this.confirmPasscode = value;
    }

    /**
     * Enables user MFA and sets view to Recovery Codes state.
     */
    public async enable(): Promise<void> {
        if (!this.confirmPasscode || this.isLoading || this.isError) return;

        this.isLoading = true;

        try {
            await this.$store.dispatch(USER_ACTIONS.ENABLE_USER_MFA, this.confirmPasscode);
            await this.$store.dispatch(USER_ACTIONS.GET);
            await this.showCodes();
            this.analytics.eventTriggered(AnalyticsEvent.MFA_ENABLED);
            await this.$notify.success('MFA was enabled successfully');
        } catch (error) {
            await this.$notify.error(error.message);
            this.isError = true;
        }

        this.isLoading = false;
    }

    /**
     * Returns satellite name from store.
     */
    private get satellite(): string {
        return this.$store.state.appStateModule.satelliteName;
    }

    /**
     * Returns pre-generated MFA secret from store.
     */
    private get userMFASecret(): string {
        return this.$store.state.usersModule.userMFASecret;
    }

    /**
     * Returns user MFA recovery codes from store.
     */
    private get userMFARecoveryCodes(): string[] {
        return this.$store.state.usersModule.userMFARecoveryCodes;
    }
}
</script>

<style scoped lang="scss">
    .enable-mfa {
        position: fixed;
        top: 0;
        bottom: 0;
        right: 0;
        left: 0;
        display: flex;
        justify-content: center;
        z-index: 1000;
        background: rgb(27 37 51 / 75%);

        &__container {
            padding: 60px;
            height: fit-content;
            margin-top: 100px;
            position: relative;
            background: #fff;
            border-radius: 6px;
            display: flex;
            flex-direction: column;
            align-items: center;
            font-family: 'font_regular', sans-serif;

            &__title {
                font-family: 'font_bold', sans-serif;
                font-size: 28px;
                line-height: 34px;
                text-align: center;
                color: #000;
                margin: 0 0 30px;
            }

            &__subtitle {
                font-size: 16px;
                line-height: 21px;
                text-align: center;
                color: #000;
                margin: 0 0 45px;
            }

            &__scan {
                padding: 25px;
                background: #f5f6fa;
                border-radius: 6px;
                display: flex;
                flex-direction: column;
                align-items: center;
                width: calc(100% - 50px);

                &__title {
                    font-family: 'font_bold', sans-serif;
                    font-size: 16px;
                    line-height: 19px;
                    text-align: center;
                    color: #000;
                    margin: 0 0 30px;
                }

                &__subtitle {
                    font-size: 14px;
                    line-height: 25px;
                    text-align: center;
                    color: #000;
                }

                &__qr {
                    margin: 30px 0;
                    background: #fff;
                    border-radius: 6px;
                    padding: 10px;

                    &__canvas {
                        height: 200px !important;
                        width: 200px !important;
                    }
                }

                &__secret {
                    margin: 5px 0 0;
                    font-family: 'font_medium', sans-serif;
                    font-size: 14px;
                    line-height: 25px;
                    text-align: center;
                    color: #000;
                }
            }

            &__confirm,
            &__codes {
                padding: 25px;
                background: #f5f6fa;
                border-radius: 6px;
                width: calc(100% - 50px);
                display: flex;
                flex-direction: column;
                align-items: center;

                &__title {
                    font-size: 16px;
                    line-height: 19px;
                    text-align: center;
                    color: #000;
                    margin-bottom: 20px;
                }
            }

            &__buttons {
                display: flex;
                align-items: center;
                width: 100%;
                margin-top: 30px;
            }

            &__close-container {
                display: flex;
                justify-content: center;
                align-items: center;
                position: absolute;
                right: 30px;
                top: 30px;
                height: 24px;
                width: 24px;
                cursor: pointer;

                &:hover .close-cross-svg-path {
                    fill: #2683ff;
                }
            }
        }
    }

    .cancel-button {
        margin-right: 15px;
    }

    .max-width {
        max-width: 485px;
    }

    @media screen and (max-height: 900px) {

        .enable-mfa {
            padding-bottom: 20px;
            overflow-y: scroll;
        }
    }
</style>
