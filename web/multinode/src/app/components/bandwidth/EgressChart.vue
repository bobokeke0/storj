// Copyright (C) 2021 Storj Labs, Inc.
// See LICENSE for copying information.

<template>
    <div class="chart">
        <p class="egress-chart__data-dimension">{{ chartDataDimension }}</p>
        <VChart
            id="egress-chart"
            :key="chartKey"
            :chart-data="chartData"
            :width="chartWidth"
            :height="chartHeight"
            :tooltip-constructor="egressTooltip"
        />
    </div>
</template>

<script lang="ts">
import { Component } from 'vue-property-decorator';

import BaseChart from '@/app/components/common/BaseChart.vue';
import VChart from '@/app/components/common/VChart.vue';

import { ChartData, Tooltip, TooltipParams, TooltipModel } from '@/app/types/chart';
import { Chart as ChartUtils } from '@/app/utils/chart';
import { BandwidthRollup } from '@/bandwidth';
import { Size } from '@/private/memory/size';

/**
 * stores egress data for egress bandwidth chart's tooltip
 */
class EgressTooltip {
    public normalEgress: string;
    public repairEgress: string;
    public auditEgress: string;
    public date: string;

    public constructor(bandwidth: BandwidthRollup) {
        this.normalEgress = Size.toBase10String(bandwidth.egress.usage);
        this.repairEgress = Size.toBase10String(bandwidth.egress.repair);
        this.auditEgress = Size.toBase10String(bandwidth.egress.audit);
        this.date = bandwidth.intervalStart.toUTCString().slice(0, 16);
    }
}

// @vue/component
@Component({
    components: {VChart},
})
export default class EgressChart extends BaseChart {
    private get allBandwidth(): BandwidthRollup[] {
        return ChartUtils.populateEmptyBandwidth(this.$store.state.bandwidth.traffic.bandwidthDaily);
    }

    public get chartDataDimension(): string {
        if (!this.$store.state.bandwidth.traffic.bandwidthDaily.length) {
            return 'Bytes';
        }

        return ChartUtils.getChartDataDimension(this.allBandwidth.map((elem) => elem.egress.audit + elem.egress.repair + elem.egress.usage));
    }

    public get chartData(): ChartData {
        let data: number[] = [0];
        const daysCount = ChartUtils.daysDisplayedOnChart();
        const chartBackgroundColor = '#edf9f4';
        const chartBorderColor = '#48a77f';
        const chartBorderWidth = 1;

        if (this.allBandwidth.length) {
            data = ChartUtils.normalizeChartData(this.allBandwidth.map(elem => elem.egress.audit + elem.egress.repair + elem.egress.usage));
        }

        return new ChartData(daysCount, chartBackgroundColor, chartBorderColor, chartBorderWidth, data);
    }

    public egressTooltip(tooltipModel: TooltipModel): void {
        const tooltipParams = new TooltipParams(tooltipModel, 'egress-chart', 'egress-tooltip',
            'egress-tooltip-point', this.tooltipMarkUp(tooltipModel),
            235, 94, 6, 4, '#48a77f');

        Tooltip.custom(tooltipParams);
    }

    private tooltipMarkUp(tooltipModel: TooltipModel): string {
        if (!tooltipModel.dataPoints) {
            return '';
        }

        const dataIndex = tooltipModel.dataPoints[0].index;
        const dataPoint = new EgressTooltip(this.allBandwidth[dataIndex]);

        return `<div class='egress-tooltip-body'>
                    <div class='egress-tooltip-body__info'>
                        <p>USAGE</p>
                        <b class="egress-tooltip-bold-text">${dataPoint.normalEgress}</b>
                    </div>
                    <div class='egress-tooltip-body__info'>
                        <p>REPAIR</p>
                        <b class="egress-tooltip-bold-text">${dataPoint.repairEgress}</b>
                    </div>
                    <div class='egress-tooltip-body__info'>
                        <p>AUDIT</p>
                        <b class="egress-tooltip-bold-text">${dataPoint.auditEgress}</b>
                    </div>
                </div>
                <div class='egress-tooltip-footer'>
                    <p>${dataPoint.date}</p>
                </div>`;
    }
}
</script>

<style lang="scss">
    .egress-chart {
        z-index: 102;

        &__data-dimension {
            font-size: 13px;
            color: var(--c-title);
            margin: 0 0 5px 31px !important;
            font-family: 'font_medium', sans-serif;
        }
    }

    #egress-tooltip {
        background: white;
        border: 1px solid var(--c-gray--light);
        min-width: 190px;
        min-height: 170px;
        font-size: 12px;
        border-radius: 14px;
        font-family: 'font_bold', sans-serif;
        color: var(--c-title);
        pointer-events: none;
        z-index: 9999;
    }

    .egress-tooltip-body {
        margin: 8px;

        &__info {
            display: flex;
            border-radius: 12px;
            padding: 14px;
            align-items: center;
            justify-content: space-between;
            margin-bottom: 14px;
            position: relative;
            font-family: 'font_bold', sans-serif;
        }
    }

    .egress-tooltip-bold-text {
        color: #2e5f46;
        font-size: 14px;
    }

    .egress-tooltip-footer {
        position: relative;
        font-size: 12px;
        width: auto;
        display: flex;
        align-items: center;
        justify-content: center;
        padding: 10px 0 16px;
        color: var(--c-title);
        font-family: 'font_bold', sans-serif;
    }
</style>
