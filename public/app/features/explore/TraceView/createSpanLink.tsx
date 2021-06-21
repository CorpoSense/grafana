import { DataLink, dateTime, Field, mapInternalLinkToExplore, rangeUtil, TimeRange } from '@grafana/data';
import { getTemplateSrv } from '@grafana/runtime';
import { Icon } from '@grafana/ui';
import { TraceSpan } from '@jaegertracing/jaeger-ui-components';
import { TraceToLogsOptions } from 'app/core/components/TraceToLogsSettings';
import { getDatasourceSrv } from 'app/features/plugins/datasource_srv';
import { SplitOpen } from 'app/types/explore';
import React from 'react';
import { LokiQuery } from '../../../plugins/datasource/loki/types';

/**
 * This is a factory for the link creator. It returns the function mainly so it can return undefined in which case
 * the trace view won't create any links and to capture the datasource and split function making it easier to memoize
 * with useMemo.
 */
export function createSpanLinkFactory(splitOpenFn: SplitOpen, traceToLogsOptions?: TraceToLogsOptions) {
  // We should return if dataSourceUid is undefined otherwise getInstanceSettings would return testDataSource.
  if (!traceToLogsOptions?.datasourceUid) {
    return undefined;
  }

  const dataSourceSettings = getDatasourceSrv().getInstanceSettings(traceToLogsOptions.datasourceUid);

  if (!dataSourceSettings) {
    return undefined;
  }

  return function (span: TraceSpan): { href: string; onClick?: (event: any) => void; content: React.ReactNode } {
    // This is reusing existing code from derived fields which may not be ideal match so some data is a bit faked at
    // the moment. Issue is that the trace itself isn't clearly mapped to dataFrame (right now it's just a json blob
    // inside a single field) so the dataLinks as config of that dataFrame abstraction breaks down a bit and we do
    // it manually here instead of leaving it for the data source to supply the config.

    const dataLink: DataLink<LokiQuery> = {
      title: dataSourceSettings.name,
      url: '',
      internal: {
        datasourceUid: dataSourceSettings.uid,
        datasourceName: dataSourceSettings.name,
        query: {
          expr: getLokiQueryFromSpan(span, traceToLogsOptions.tags),
          refId: '',
        },
      },
    };

    const link = mapInternalLinkToExplore({
      link: dataLink,
      internalLink: dataLink.internal!,
      scopedVars: {},
      range: getTimeRangeFromSpan(span, traceToLogsOptions),
      field: {} as Field,
      onClickFn: splitOpenFn,
      replaceVariables: getTemplateSrv().replace.bind(getTemplateSrv()),
    });

    return {
      href: link.href,
      onClick: link.onClick,
      content: <Icon name="gf-logs" title="Explore the logs for this in split view" />,
    };
  };
}

/**
 * Default keys to use when there are no configured tags.
 */
const defaultKeys = ['cluster', 'hostname', 'namespace', 'pod'];

function getLokiQueryFromSpan(span: TraceSpan, keys?: string[]): string {
  const keysToCheck = keys?.length ? keys : defaultKeys;
  const tags = [...span.process.tags, ...span.tags].reduce((acc, tag) => {
    if (keysToCheck.includes(tag.key)) {
      acc.push(`${tag.key}="${tag.value}"`);
    }
    return acc;
  }, [] as string[]);
  return `{${tags.join(', ')}}`;
}

/**
 * Gets a time range from the span.
 */
function getTimeRangeFromSpan(span: TraceSpan, traceToLogsOptions?: TraceToLogsOptions): TimeRange {
  const adjustedStartTime = traceToLogsOptions?.spanStartTimeShift
    ? span.startTime / 1000 + rangeUtil.intervalToMs(traceToLogsOptions.spanStartTimeShift)
    : span.startTime / 1000;
  const from = dateTime(adjustedStartTime);
  const spanEndMs = (span.startTime + span.duration) / 1000;
  const adjustedEndTime = traceToLogsOptions?.spanEndTimeShift
    ? spanEndMs + rangeUtil.intervalToMs(traceToLogsOptions.spanEndTimeShift)
    : spanEndMs;
  const to = dateTime(adjustedEndTime);

  return {
    from,
    to,
    // Weirdly Explore does not handle ISO string which would have been the default stringification if passed as object
    // and we have to use this custom format :( .
    raw: {
      from: from.utc().format('YYYYMMDDTHHmmss'),
      to: to.utc().format('YYYYMMDDTHHmmss'),
    },
  };
}
