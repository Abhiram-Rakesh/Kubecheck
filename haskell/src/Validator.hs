{-# LANGUAGE OverloadedStrings #-}

module Validator
    ( validateResource
    , extractContainers
    , K8sResource(..)   -- ← Re-exported
    , Violation(..)     -- ← Re-exported
    ) where

import Rules
import Types

-- | Validate a Kubernetes resource and return all violations
validateResource :: K8sResource -> [Violation]
validateResource resource =
    let resourceContainers = extractContainers resource
    in concatMap validateContainer resourceContainers

-- | Extract containers from a Kubernetes resource
-- Handles Deployment, StatefulSet, DaemonSet, ReplicaSet, Job, CronJob, Pod
extractContainers :: K8sResource -> [Container]
extractContainers resource =
    case spec resource of
        Nothing -> []
        Just s  -> extractFromSpec s
  where
    extractFromSpec :: Spec -> [Container]
    extractFromSpec s =
        case template s of
            -- Workload with template (Deployment, StatefulSet, etc.)
            Just tmpl -> podContainers (podSpec tmpl)
            -- Direct pod spec (Pod)
            Nothing -> case containers s of
                Just cs -> cs
                Nothing -> []

-- | Validate a single container using all rules
validateContainer :: Container -> [Violation]
validateContainer container =
    concat [ checkNoLatestImage container
           , checkRequireResources container
           , checkNoRootContainers container
           ]

